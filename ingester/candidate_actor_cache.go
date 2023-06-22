package ingester

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	bff "github.com/strideynet/bsky-furry-feed"
	"github.com/strideynet/bsky-furry-feed/store"
	"go.uber.org/zap"
	"sync"
	"time"
)

// CandidateActorCache holds a view of the candidate actors from
// the database, refreshing itself every minute. It's designed to be safely
// called concurrently. This prevents us needing to hit the database for every
// event which would produce significant load on the db and also increase the
// amount of time it takes to handle an event we aren't interested in.
// The only downside to this approach is that it takes up to a minute for
// new candidate repositories to be monitored.
type CandidateActorCache struct {
	log     *zap.Logger
	queries *store.Queries

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
	cached map[string]bff.CandidateActor
	// mu protects cached to prevent concurrent access leading to corruption.
	mu sync.RWMutex
}

func NewCandidateActorCache(
	log *zap.Logger,
	queries *store.Queries,
) *CandidateActorCache {
	return &CandidateActorCache{
		queries:        queries,
		log:            log,
		period:         time.Minute,
		refreshTimeout: time.Second * 10,
	}
}

func (crc *CandidateActorCache) GetByDID(
	did string,
) *bff.CandidateActor {
	crc.mu.RLock()
	defer crc.mu.RUnlock()
	v, ok := crc.cached[did]
	if ok {
		return &v
	}
	return nil
}

func (crc *CandidateActorCache) Sync(ctx context.Context) error {
	crc.log.Info("starting cache sync")
	data, err := crc.queries.ListCandidateActors(ctx, store.NullActorStatus{
		Valid: false,
	})
	if err != nil {
		return fmt.Errorf("listing candidate actors: %w", err)
	}

	mapped := map[string]bff.CandidateActor{}
	for _, cr := range data {
		mapped[cr.DID] = bff.CandidateActorFromStore(cr)
	}

	crc.mu.Lock()
	defer crc.mu.Unlock()
	crc.cached = mapped
	crc.log.Info("finished cache sync", zap.Int("count", len(mapped)))
	return nil
}

func (crc *CandidateActorCache) Start(ctx context.Context) error {
	ticker := time.NewTicker(crc.period)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
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

func (crc *CandidateActorCache) CreatePendingCandidateActor(ctx context.Context, did string) (*bff.CandidateActor, error) {
	ctx, span := tracer.Start(ctx, "candidate_actor_cache.create_pending_candidate_actor")
	defer span.End()
	params := store.CreateCandidateActorParams{
		DID:     did,
		Comment: "added by system",
		CreatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
		Status: store.ActorStatusPending,
	}
	ca, err := crc.queries.CreateCandidateActor(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("creating candidate actor: %w", err)
	}
	crc.log.Info("added new pending actor")

	crc.mu.Lock()
	defer crc.mu.Unlock()
	converted := bff.CandidateActorFromStore(ca)
	crc.cached[ca.DID] = converted
	return &converted, nil
}
