package ingester

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jonboulle/clockwork"

	v1 "github.com/strideynet/bsky-furry-feed/proto/bff/v1"
	"github.com/strideynet/bsky-furry-feed/store"
	"go.uber.org/zap"
)

// ActorCache holds a view of the candidate actors from
// the database, refreshing itself every minute. It's designed to be safely
// called concurrently. This prevents us needing to hit the database for every
// event which would produce significant load on the db and also increase the
// amount of time it takes to handle an event we aren't interested in.
// The only downside to this approach is that it takes up to a minute for
// new candidate repositories to be monitored.
//
// TODO: Move this to the store as a wrapper around *store.PGXStore ?
type ActorCache struct {
	log   *zap.Logger
	store *store.PGXStore

	clock clockwork.Clock

	// period is how often to attempt to fresh the list of candidate
	// actors.
	period time.Duration
	// refreshTimeout is how long to give any attempt to complete. This is
	// necessary to prevent a hung iteration from blocking the loop.
	// Realistically, we don't expect this process to take any longer than
	// ten seconds.
	refreshTimeout time.Duration

	// cached is a map keyed by the actor DID to the data about the
	// actor. The go standard map implementation is fast enough for our
	// needs at this time.
	cached map[string]*v1.Actor
	// mu protects cached to prevent concurrent access leading to corruption.
	mu sync.RWMutex
}

func NewActorCache(
	log *zap.Logger,
	store *store.PGXStore,
) *ActorCache {
	return &ActorCache{
		store:          store,
		clock:          clockwork.NewRealClock(),
		log:            log,
		period:         time.Minute,
		refreshTimeout: time.Second * 10,
	}
}

func (crc *ActorCache) GetByDID(
	did string,
) *v1.Actor {
	crc.mu.RLock()
	defer crc.mu.RUnlock()
	v, ok := crc.cached[did]
	if ok {
		return v
	}
	return nil
}

func (crc *ActorCache) Sync(ctx context.Context) error {
	crc.log.Info("starting cache sync")
	data, err := crc.store.ListActors(ctx, store.ListActorsOpts{})
	if err != nil {
		return fmt.Errorf("listing actors: %w", err)
	}

	mapped := map[string]*v1.Actor{}
	for _, cr := range data {
		mapped[cr.Did] = cr
	}

	crc.mu.Lock()
	defer crc.mu.Unlock()
	crc.cached = mapped
	crc.log.Info("finished cache sync", zap.Int("count", len(mapped)))
	return nil
}

func (crc *ActorCache) Start(ctx context.Context) error {
	ticker := crc.clock.NewTicker(crc.period)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.Chan():
			// TODO: If this fails enough times, we should bail out, this allows
			// a process restart to potentially rectify the situation.
			ctx, cancel := context.WithTimeout(ctx, crc.refreshTimeout)
			if err := crc.Sync(ctx); err != nil {
				crc.log.Error("failed to fill cache", zap.Error(err))
			}
			cancel()
		}
	}
}

func (crc *ActorCache) CreatePendingCandidateActor(ctx context.Context, did string) (err error) {
	ctx, span := tracer.Start(ctx, "actor_cache.create_pending_actor")
	defer func() {
		endSpan(span, err)
	}()
	params := store.CreateActorOpts{
		DID:     did,
		Comment: "added by system",
		Status:  v1.ActorStatus_ACTOR_STATUS_PENDING,
	}
	ca, err := crc.store.CreateActor(ctx, params)
	if err != nil {
		return fmt.Errorf("creating actor: %w", err)
	}
	crc.log.Info("added new pending actor")

	crc.mu.Lock()
	defer crc.mu.Unlock()
	crc.cached[ca.Did] = ca
	return nil
}

func (crc *ActorCache) OptInOrMarkPending(ctx context.Context, did string) (err error) {
	ctx, span := tracer.Start(ctx, "actor_cache.opt_in")
	defer func() {
		endSpan(span, err)
	}()

	status, err := crc.store.OptInOrMarkActorPending(ctx, did)
	if err != nil {
		return fmt.Errorf("opting in actor: %w", err)
	}

	crc.mu.Lock()
	defer crc.mu.Unlock()
	ca := crc.cached[did]
	if ca != nil {
		ca.Status = status
	}

	return nil
}

func (crc *ActorCache) OptOutOrForget(ctx context.Context, did string) (err error) {
	ctx, span := tracer.Start(ctx, "actor_cache.opt_out")
	defer func() {
		endSpan(span, err)
	}()

	status, err := crc.store.OptOutOrForgetActor(ctx, did)
	if err != nil {
		return fmt.Errorf("opting out actor: %w", err)
	}

	crc.mu.Lock()
	defer crc.mu.Unlock()
	if status == v1.ActorStatus_ACTOR_STATUS_NONE {
		delete(crc.cached, did)
	}

	return nil
}
