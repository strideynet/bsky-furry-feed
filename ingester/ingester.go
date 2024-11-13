package ingester

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/bluesky-social/indigo/events/schedulers/sequential"

	"net/http"
	"sync"
	"time"

	"github.com/bluesky-social/indigo/util"

	jsclient "github.com/bluesky-social/jetstream/pkg/client"
	jsparallel "github.com/bluesky-social/jetstream/pkg/client/schedulers/parallel"
	"github.com/bluesky-social/jetstream/pkg/models"

	v1 "github.com/strideynet/bsky-furry-feed/proto/bff/v1"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/events"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/strideynet/bsky-furry-feed/store"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// tracer is the BFF wide tracer. This is different to the tracer used in
// the feedIngester which has a different sampling rate.
var tracer = otel.Tracer("github.com/strideynet/bsky-furry-feed/ingester")

var workItemsProcessed = promauto.NewSummaryVec(prometheus.SummaryOpts{
	Name: "bff_ingester_work_item_duration_seconds",
	Help: "The total number of work items handled by the ingester worker pool.",
}, []string{"type"})

type actorCacher interface {
	GetByDID(did string) *v1.Actor
	CreatePendingCandidateActor(ctx context.Context, did string) (err error)
}

var workerCursors = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "bff_ingester_worker_cursors",
	Help: "The current cursor a worker is at.",
}, []string{"worker"})

var flushedWorkerCursor = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "bff_ingester_flushed_worker_cursor",
	Help: "The current cursor flushed to persistent storage.",
})

type FirehoseIngester struct {
	// dependencies
	log        *zap.Logger
	actorCache actorCacher
	store      *store.PGXStore

	// configuration
	subscribeURL        string
	workerCount         int
	workItemTimeout     time.Duration
	cursorFlushInterval time.Duration
}

// TODO: Eventually make this a worker struct.
// TODO: Capture a worker "status" e.g idle/working
type workerState struct {
	mu               sync.Mutex
	lastProcessedSeq int64
}

func NewFirehoseIngester(
	log *zap.Logger, store *store.PGXStore, crc *ActorCache, bgsHost string,
) *FirehoseIngester {
	return &FirehoseIngester{
		log:        log,
		actorCache: crc,
		store:      store,

		subscribeURL:        bgsHost + "/xrpc/com.atproto.sync.subscribeRepos",
		workerCount:         8,
		workItemTimeout:     time.Second * 30,
		cursorFlushInterval: time.Second * 10,
	}
}

func (fi *FirehoseIngester) Start(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)

	// Unbuffered channel so that the websocket will stop reading if the workers
	// are not ready. In future, we may want to consider some reasonable
	// buffering to account for short spikes in event rates.
	evtChan := make(chan *atproto.SyncSubscribeRepos_Commit)

	workerStates := make([]*workerState, fi.workerCount)
	flushCommitCursor := func(ctx context.Context) error {
		// Determine the lowest last processed seq of all the workers. We want
		// to persist the lowest last processed seq to ensure we never miss
		// data, but this may cause some reprocessing.
		var lowestLastSeq int64 = -1
		for _, w := range workerStates {
			w.mu.Lock()
			lastSeq := w.lastProcessedSeq
			w.mu.Unlock()
			// Worker hasn't processed it's first commit yet, so we ignore it.
			if lastSeq == -1 {
				continue
			}
			if lowestLastSeq == -1 || lastSeq < lowestLastSeq {
				lowestLastSeq = lastSeq
			}
		}

		if lowestLastSeq == -1 {
			return fmt.Errorf("no workers reported work, cannot persist cursor")
		}

		if err := fi.store.SetFirehoseCommitCursor(ctx, lowestLastSeq); err != nil {
			return fmt.Errorf("saving cursor: %w", err)
		}
		fi.log.Info("successfully flushed cursor", zap.Int64("cursor", lowestLastSeq))
		flushedWorkerCursor.Set(float64(lowestLastSeq))
		return nil
	}
	defer func() {
		// Flush the commit cursor one final time on shutdown.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		if err := flushCommitCursor(ctx); err != nil {
			fi.log.Error(
				"failed to flush final firehose commit cursor",
				zap.Error(err),
			)
		}
	}()

	eg.Go(func() error {
		workerWg := sync.WaitGroup{}

		for i := 0; i < fi.workerCount; i++ {
			i := i
			state := &workerState{
				lastProcessedSeq: -1,
			}
			workerStates[i] = state

			workerWg.Add(1)
			go func() {
				defer workerWg.Done()
				for {
					select {
					case <-ctx.Done():
						fi.log.Warn("worker exiting", zap.Int("worker", i))
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

						state.mu.Lock()
						if evt.Seq <= state.lastProcessedSeq {
							fi.log.Error(
								"cursor went backwards or was repeated",
								zap.Int64("seq", evt.Seq),
								zap.Int64("cursor", state.lastProcessedSeq),
							)
						}
						state.lastProcessedSeq = evt.Seq
						state.mu.Unlock()

						workerCursors.WithLabelValues(strconv.Itoa(i)).Set(float64(evt.Seq))
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

		for {
			select {
			case <-ctx.Done():
				fi.log.Warn("cursor flushing worker exiting")
				return nil
			case <-time.After(fi.cursorFlushInterval):
			}

			if err := flushCommitCursor(ctx); err != nil {
				fi.log.Error(
					"failed to flush firehose commit cursor",
					zap.Error(err),
				)
			}
		}
	})

	eg.Go(func() error {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		cursor, err := fi.store.GetFirehoseCommitCursor(ctx)
		if err != nil {
			return fmt.Errorf("get commit cursor: %w", err)
		}
		fi.log.Info("starting ingestion", zap.Int64("cursor", cursor))

		subscribeURL := fi.subscribeURL
		if cursor != -1 {
			subscribeURL += "?cursor=" + strconv.FormatInt(cursor, 10)
		}

		con, _, err := websocket.DefaultDialer.DialContext(
			ctx, subscribeURL, http.Header{},
		)
		if err != nil {
			return fmt.Errorf("dialing websocket: %w", err)
		}

		// TODO: Indigo now offers a native parallel consumer pool, we should
		// consider switching to it - but only if we can
		callbacks := &events.RepoStreamCallbacks{
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
		}
		scheduler := sequential.NewScheduler("main", callbacks.EventHandler)
		if err := events.HandleRepoStream(ctx, con, scheduler); err != nil {
			fi.log.Error("repo stream from relay has failed", zap.Error(err))
			return fmt.Errorf("handling repo stream: %w", err)
		}
		return nil
		// TODO: sometimes stream exits of own accord, we should attempt to
		// reconnect and enter an "error state".
	})

	return eg.Wait()
}

func (fi *FirehoseIngester) StartJetstreamConsumption(ctx context.Context) (err error) {
	eg, ctx := errgroup.WithContext(ctx)

	jsCfg := jsclient.DefaultClientConfig()
	// TODO: Setup config

	sched := jsparallel.NewScheduler(
		fi.workerCount,
		"jetstream",
		slog.Default(), // TODO: Switch from Zap to Slog,
		func(ctx context.Context, e *models.Event) error {
			ctx, cancel := context.WithTimeout(ctx, fi.workItemTimeout)
			defer cancel()

			// Ignore events other than commit.
			if e.Commit == nil {
				return nil
			}

			if err := fi.handleCommit(ctx, e); err != nil {
				fi.log.Error(
					"failed to handle commit",
					zap.Error(err),
					zap.Any("evt", e),
				)
				return fmt.Errorf("handling commit: %w", err)
			}

			// Persist cursor
		},
	)

	jsClient, err := jsclient.NewClient(
		jsCfg, slog.Default(), sched,
	)
	if err != nil {
		return fmt.Errorf("creating jetstream client: %w", err)
	}

	// fetch cursor and set back 15 minutes, or default to now.

	eg.Go(func() error {
		if err := jsClient.ConnectAndRead(ctx, nil); err != nil {
			return fmt.Errorf("reading jetstream: %w", err)
		}
		return nil
	})

	eg.Go(func() error {
		// TODO: Update persisted cursor....
	})

	return eg.Wait()
}

func (fi *FirehoseIngester) handleCommit(
	ctx context.Context, evt *models.Event,
) (err error) {
	ctx, span := tracer.Start(ctx, "firehose_ingester.handle_commit")
	defer func() {
		endSpan(span, err)
	}()
	span.SetAttributes(actorDIDAttr(evt.Did))

	evtTime := time.UnixMicro(evt.TimeUS)
	uri := fmt.Sprintf(
		"at://%s/%s/%s",
		evt.Did, evt.Commit.Collection, evt.Commit.RKey,
	)

	switch evt.Commit.Operation {
	case models.CommitOperationCreate:
		if err := fi.handleRecordCreate(
			ctx,
			evt.Did,
			uri,
			evt.Commit.Collection,
			evt.Commit.Record,
		); err != nil {
			return fmt.Errorf("create (%s): handling record create: %w", uri, err)
		}
	case models.CommitOperationUpdate:
		if err := fi.handleRecordUpdate(
			ctx,
			evt.Did,
			evt.Commit.Rev,
			uri,
			evtTime,
			evt.Commit.Collection,
			evt.Commit.Record,
		); err != nil {
			return fmt.Errorf("update (%s): handling record update: %w", uri, err)
		}
	case models.CommitOperationDelete:
		if err := fi.handleRecordDelete(
			ctx, evt.Did, uri,
		); err != nil {
			return fmt.Errorf("handling record delete: %w", err)
		}
	default:
		fi.log.Warn("unknown commit operation", zap.String("kind", evt.Kind))
	}

	return nil
}

func endSpan(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
}

func (fi *FirehoseIngester) handleRecordCreate(
	ctx context.Context,
	repoDID string,
	recordUri string,
	recordCollection string,
	record json.RawMessage,
) (err error) {
	ctx, span := tracer.Start(ctx, "firehose_ingester.handle_record_create")
	defer func() {
		endSpan(span, err)
	}()
	span.SetAttributes(recordUriAttr(recordUri))

	actor := fi.actorCache.GetByDID(repoDID)
	if actor == nil {
		// Check if it's a follow of the furry feed
		if recordCollection != "app.bsky.graph.follow" {
			return nil
		}
		data := &bsky.GraphFollow{}
		if err := json.Unmarshal(record, data); err != nil {
			return fmt.Errorf("unmarshalling record: %w", err)
		}
		// If it's an unknown actor, and they've interacted, add em to
		// the candidate actor store with pending status. Otherwise, ignore
		// them.
		// TODO: Make this not hard coded
		// https://bsky.app/profile/furryli.st
		if data.Subject != "did:plc:jdkvwye2lf4mingzk7qdebzc" {
			return nil
		}
		fi.log.Info(
			"unknown actor interacted, adding to db as pending",
			zap.String("did", repoDID),
		)
		if err := fi.actorCache.CreatePendingCandidateActor(ctx, repoDID); err != nil {
			return fmt.Errorf("creating pending candidate actor: %w", err)
		}

		return nil
	}

	// Only collect events from actors we care about e.g those that are
	// approved.
	if !(actor.Status == v1.ActorStatus_ACTOR_STATUS_APPROVED) {
		return nil
	}

	switch recordCollection {
	case "app.bsky.feed.post":
		data := &bsky.FeedPost{}
		if err := json.Unmarshal(record, data); err != nil {
			return fmt.Errorf("unmarshalling record: %w", err)
		}
		err := fi.handleFeedPostCreate(ctx, repoDID, recordUri, data)
		if err != nil {
			return fmt.Errorf("handling app.bsky.feed.post create: %w", err)
		}
	case "app.bsky.feed.like":
		data := &bsky.FeedLike{}
		if err := json.Unmarshal(record, data); err != nil {
			return fmt.Errorf("unmarshalling record: %w", err)
		}
		err := fi.handleFeedLikeCreate(ctx, repoDID, recordUri, data)
		if err != nil {
			return fmt.Errorf("handling app.bsky.feed.like: %w", err)
		}
	case "app.bsky.graph.follow":
		data := &bsky.GraphFollow{}
		if err := json.Unmarshal(record, data); err != nil {
			return fmt.Errorf("unmarshalling record: %w", err)
		}
		err := fi.handleGraphFollowCreate(ctx, repoDID, recordUri, data)
		if err != nil {
			return fmt.Errorf("handling app.bsky.graph.follow: %w", err)
		}
	default:
		span.AddEvent("ignoring record due to unrecognized type")
	}

	return nil
}

func actorDIDAttr(s string) attribute.KeyValue {
	return attribute.String("actor.did", s)
}

func recordUriAttr(s string) attribute.KeyValue {
	return attribute.String("record.uri", s)
}

func (fi *FirehoseIngester) handleRecordDelete(
	ctx context.Context,
	repoDID string,
	recordUri string,
) (err error) {
	ctx, span := tracer.Start(ctx, "firehose_ingester.handle_record_delete")
	defer func() {
		endSpan(span, err)
	}()
	span.SetAttributes(recordUriAttr(recordUri))

	actor := fi.actorCache.GetByDID(repoDID)
	if actor == nil {
		// if we don’t know the actor, we don’t have their data
		return
	}

	parsedUri, err := util.ParseAtUri(recordUri)
	if err != nil {
		return fmt.Errorf("parsing uri: %w", err)
	}

	switch parsedUri.Collection {
	case "app.bsky.actor.profile":
		err = fi.handleActorProfileDelete(ctx, recordUri)
	case "app.bsky.feed.post":
		err = fi.handleFeedPostDelete(ctx, recordUri)
	case "app.bsky.feed.like":
		err = fi.handleFeedLikeDelete(ctx, recordUri)
	case "app.bsky.graph.follow":
		err = fi.handleGraphFollowDelete(ctx, recordUri)
	default:
		span.AddEvent("ignoring record due to unrecognized type")
	}

	return
}

func (fi *FirehoseIngester) handleRecordUpdate(
	ctx context.Context,
	repoDID string,
	repoRev string,
	recordUri string,
	updatedAt time.Time,
	recordCollection string,
	record json.RawMessage,
) (err error) {
	ctx, span := tracer.Start(ctx, "firehose_ingester.handle_record_update")
	defer func() {
		endSpan(span, err)
	}()
	span.SetAttributes(recordUriAttr(recordUri))

	actor := fi.actorCache.GetByDID(repoDID)
	if actor == nil {
		return nil
	}

	// Only collect events from actors we care about e.g those that are
	// approved.
	if !(actor.Status == v1.ActorStatus_ACTOR_STATUS_APPROVED) {
		return nil
	}

	switch recordCollection {
	case "app.bsky.actor.profile":
		data := &bsky.ActorProfile{}
		if err := json.Unmarshal(record, data); err != nil {
			return fmt.Errorf("unmarshalling record: %w", err)
		}
		err := fi.handleActorProfileUpdate(ctx, repoDID, repoRev, recordUri, updatedAt, data)
		if err != nil {
			return fmt.Errorf("handling app.bsky.actor.profile update: %w", err)
		}
	default:
		span.AddEvent("ignoring record due to unrecognized type")
	}

	return nil
}
