package ingester

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/events"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/repo"
	"github.com/bluesky-social/indigo/repomgr"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/strideynet/bsky-furry-feed/store"
	typegen "github.com/whyrusleeping/cbor-gen"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"golang.org/x/sync/errgroup"
	"net/http"
	"sync"
	"time"
)

var tracer = otel.Tracer("github.com/strideynet/bsky-furry-feed/ingester")

var workItemsProcessed = promauto.NewSummaryVec(prometheus.SummaryOpts{
	Name: "bff_ingester_work_item_duration_seconds",
	Help: "The total number of work items handled by the ingester worker pool.",
}, []string{"type"})

type FirehoseIngester struct {
	// dependencies
	log     *zap.Logger
	crc     *CandidateActorCache
	queries *store.Queries

	// configuration
	subscribeURL    string
	workerCount     int
	workItemTimeout time.Duration

	furryFeeds []string
}

func NewFirehoseIngester(
	log *zap.Logger, queries *store.Queries, crc *CandidateActorCache,
) *FirehoseIngester {
	return &FirehoseIngester{
		log:     log,
		crc:     crc,
		queries: queries,

		subscribeURL:    "wss://bsky.social/xrpc/com.atproto.sync.subscribeRepos",
		workerCount:     6,
		workItemTimeout: time.Second * 30,
		furryFeeds: []string{
			"at://did:plc:jdkvwye2lf4mingzk7qdebzc/app.bsky.feed.generator/furry-new",
			"at://did:plc:jdkvwye2lf4mingzk7qdebzc/app.bsky.feed.generator/furry-hot",
			"at://did:plc:jdkvwye2lf4mingzk7qdebzc/app.bsky.feed.generator/furry-fursuit",
			"at://did:plc:jdkvwye2lf4mingzk7qdebzc/app.bsky.feed.generator/furry-art",
			"at://did:plc:jdkvwye2lf4mingzk7qdebzc/app.bsky.feed.generator/furry-nsfw",
		},
	}
}

func (fi *FirehoseIngester) Start(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)

	// Unbuffered channel so that the websocket will stop reading if the workers
	// are not ready. In future, we may want to consider some reasonable
	// buffering to account for short spikes in event rates.
	evtChan := make(chan *atproto.SyncSubscribeRepos_Commit)
	eg.Go(func() error {
		workerWg := sync.WaitGroup{}
		for n := 1; n < fi.workerCount; n++ {
			n := n
			workerWg.Add(1)
			go func() {
				defer workerWg.Done()
				for {
					select {
					case <-ctx.Done():
						fi.log.Warn("worker exiting", zap.Int("worker", n))
						return
					case evt := <-evtChan:
						// record start time so we can collect
						start := time.Now()
						// 30 seconds max to deal with any work item. This
						// prevents a worker hanging.
						ctx, cancel := context.WithTimeout(
							ctx, fi.workItemTimeout,
						)
						if err := fi.handleCommit(ctx, evt); err != nil {
							fi.log.Error(
								"failed to handle repo commit",
								zap.Error(err),
							)
						}
						workItemsProcessed.
							WithLabelValues("repo_commit").
							Observe(time.Since(start).Seconds())
						cancel()
					}
				}
			}()
		}
		workerWg.Wait()
		return nil
	})

	eg.Go(func() error {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		con, _, err := websocket.DefaultDialer.DialContext(
			ctx, fi.subscribeURL, http.Header{},
		)
		if err != nil {
			return fmt.Errorf("dialing websocket: %w", err)
		}

		go func() {
			<-ctx.Done()
			fi.log.Warn("closing websocket subscription")
			if err := con.Close(); err != nil {
				fi.log.Error(
					"error occurred closing websocket",
					zap.Error(err),
				)
			}
			fi.log.Warn("closed websocket subscription")
		}()
		// TODO: sometimes stream exits of own accord, we should attempt to
		// reconnect several times and then return an error to cause the
		// process to crash out.
		return events.HandleRepoStream(ctx, con, &events.RepoStreamCallbacks{
			RepoCommit: func(evt *atproto.SyncSubscribeRepos_Commit) error {
				select {
				case <-ctx.Done():
					// Ensure we don't get stuck waiting for a worker even if
					// the connection has shutdown.
					return nil
				case evtChan <- evt:
				}
				return nil
			},
		})
	})

	return eg.Wait()
}

func (fi *FirehoseIngester) handleCommit(ctx context.Context, evt *atproto.SyncSubscribeRepos_Commit) error {
	// TODO: Find a way to use tail-based sampling so that we can capture this trace
	// before candidateActor is run and ensure we always capture candidateActor
	// traces.
	ctx, span := tracer.Start(ctx, "firehose_ingester.handle_commit")
	defer span.End()
	span.SetAttributes(
		attribute.String("candidate_actor.did", evt.Repo),
	)
	log := fi.log.With(zap.String("candidate_actor.did", evt.Repo))
	rr, err := repo.ReadRepoFromCar(ctx, bytes.NewReader(evt.Blocks))
	if err != nil {
		return fmt.Errorf("reading repo from car %w", err)
	}
	for _, op := range evt.Ops {
		// Ignore any op that isn't a record create.
		if repomgr.EventKind(op.Action) != repomgr.EvtKindCreateRecord {
			continue
		}
		log := log.With(
			zap.String("op.action", op.Action),
			zap.String("op.path", op.Path),
		)

		recordCid, record, err := rr.GetRecord(ctx, op.Path)
		if err != nil {
			if errors.Is(err, lexutil.ErrUnrecognizedType) {
				continue
			}
			return fmt.Errorf("getting record for op: %w", err)
		}
		// Ensure there isn't a mismatch between the reference and the found
		// object.
		if lexutil.LexLink(recordCid) != *op.Cid {
			return fmt.Errorf("mismatch in record and op cid: %s != %s", recordCid, *op.Cid)
		}

		uri := fmt.Sprintf("at://%s/%s", evt.Repo, op.Path)
		log = log.With(zap.String("record.uri", uri))
		if err := fi.handleRecordCreate(
			ctx, log, evt.Repo, uri, record,
		); err != nil {
			return fmt.Errorf("handling record create: %w", err)
		}
	}

	return nil
}

func (fi *FirehoseIngester) isFurryFeedLike(record typegen.CBORMarshaler) bool {
	like, ok := record.(*bsky.FeedLike)
	if !ok {
		return false
	}

	return slices.Contains(fi.furryFeeds, like.Subject.Uri)
}

func (fi *FirehoseIngester) handleRecordCreate(
	ctx context.Context,
	log *zap.Logger,
	repoDID string,
	recordUri string,
	record typegen.CBORMarshaler,
) error {
	ctx, span := tracer.Start(ctx, "firehose_ingester.handle_record_create")
	defer span.End()

	actor := fi.crc.GetByDID(repoDID)
	if actor == nil {
		// If it's an unknown actor, and they've just liked the feed, add em to
		// the candidate actor store with pending status. Otherwise, ignore
		// them.
		if !fi.isFurryFeedLike(record) {
			return nil
		}
		log.Info("unknown actor liked feed, adding to db as pending", zap.String("did", repoDID))
		var err error
		actor, err = fi.crc.CreatePendingCandidateActor(ctx, repoDID)
		if err != nil {
			return fmt.Errorf("creating pending candidate actor: %w", err)
		}
	}

	// Only collect events from actors we care about e.g those that are
	// approved or pending. Collecting pending actor's data means as soon as
	// they are approved, their data can contribute to the feeds immediately.
	if !(actor.Status == store.ActorStatusApproved || actor.Status == store.ActorStatusPending) {
		return nil
	}

	log.Debug("handling record create", zap.Any("record", record))
	switch data := record.(type) {
	case *bsky.FeedPost:
		err := fi.handleFeedPostCreate(ctx, log, repoDID, recordUri, data)
		if err != nil {
			return fmt.Errorf("handling app.bsky.feed.post create: %w", err)
		}
	case *bsky.FeedLike:
		err := fi.handleFeedLikeCreate(ctx, log, repoDID, recordUri, data)
		if err != nil {
			return fmt.Errorf("handling app.bsky.feed.like: %w", err)
		}
	case *bsky.GraphFollow:
		err := fi.handleGraphFollowCreate(ctx, log, repoDID, recordUri, data)
		if err != nil {
			return fmt.Errorf("handling app.bsky.graph.follow: %w", err)
		}
	default:
		log.Debug("ignoring record create due to handled type")
	}

	return nil
}
