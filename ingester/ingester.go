package ingester

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/bluesky-social/indigo/events/schedulers/sequential"

	"github.com/bluesky-social/indigo/util"
	"github.com/ipfs/go-cid"

	"net/http"
	"sync"
	"time"

	"github.com/strideynet/bsky-furry-feed/bluesky"
	v1 "github.com/strideynet/bsky-furry-feed/proto/bff/v1"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

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
	OptInOrMarkPending(ctx context.Context, did string) (err error)
	OptOutOrForget(ctx context.Context, did string) (err error)
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
	log *zap.Logger, store *store.PGXStore, crc *ActorCache, pdsHost string,
) *FirehoseIngester {
	return &FirehoseIngester{
		log:        log,
		actorCache: crc,
		store:      store,

		subscribeURL:        pdsHost + "/xrpc/com.atproto.sync.subscribeRepos",
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
		return events.HandleRepoStream(ctx, con, scheduler)
		// TODO: sometimes stream exits of own accord, we should attempt to
		// reconnect and enter an "error state".
	})

	return eg.Wait()
}

func (fi *FirehoseIngester) handleCommit(ctx context.Context, evt *atproto.SyncSubscribeRepos_Commit) (err error) {
	ctx, span := tracer.Start(ctx, "firehose_ingester.handle_commit")
	defer func() {
		endSpan(span, err)
	}()
	span.SetAttributes(actorDIDAttr(evt.Repo))

	commitCID := cid.Cid(evt.Commit)

	time, err := bluesky.ParseTime(evt.Time)
	if err != nil {
		return fmt.Errorf("parsing timestamp: %w", err)
	}
	rr, err := repo.ReadRepoFromCar(ctx, bytes.NewReader(evt.Blocks))
	if err != nil {
		return fmt.Errorf("reading repo from car: %w", err)
	}
	for _, op := range evt.Ops {
		uri := fmt.Sprintf("at://%s/%s", evt.Repo, op.Path)

		switch repomgr.EventKind(op.Action) {
		case repomgr.EvtKindCreateRecord:
			recordCid, record, err := rr.GetRecord(ctx, op.Path)
			if err != nil {
				if errors.Is(err, lexutil.ErrUnrecognizedType) {
					continue
				}
				return fmt.Errorf("create (%s): getting record for op: %w", uri, err)
			}

			// Ensure there isn't a mismatch between the reference and the found
			// object.
			if lexutil.LexLink(recordCid) != *op.Cid {
				return fmt.Errorf("create (%s): mismatch in record and op cid: %s != %s", uri, recordCid, *op.Cid)
			}

			if err := fi.handleRecordCreate(
				ctx, evt.Repo, uri, record,
			); err != nil {
				return fmt.Errorf("create (%s): handling record create: %w", uri, err)
			}
		case repomgr.EvtKindUpdateRecord:
			recordCid, record, err := rr.GetRecord(ctx, op.Path)
			if err != nil {
				if errors.Is(err, lexutil.ErrUnrecognizedType) {
					continue
				}
				return fmt.Errorf("update (%s): getting record for op: %w", uri, err)
			}

			// Ensure there isn't a mismatch between the reference and the found
			// object.
			if lexutil.LexLink(recordCid) != *op.Cid {
				return fmt.Errorf("update (%s): mismatch in record and op cid: %s != %s", uri, recordCid, *op.Cid)
			}

			if err := fi.handleRecordUpdate(
				ctx, evt.Repo, commitCID, uri, time, record,
			); err != nil {
				return fmt.Errorf("update (%s): handling record update: %w", uri, err)
			}
		case repomgr.EvtKindDeleteRecord:
			if err := fi.handleRecordDelete(
				ctx, evt.Repo, uri,
			); err != nil {
				return fmt.Errorf("handling record delete: %w", err)
			}
		}
	}

	return nil
}

func (fi *FirehoseIngester) IsFurryFeedDID(did string) bool {
	// TODO: Make this not hard coded
	// https://bsky.app/profile/furryli.st
	return did == "did:plc:jdkvwye2lf4mingzk7qdebzc"
}

func (fi *FirehoseIngester) isFurryFeedFollow(record typegen.CBORMarshaler) bool {
	follow, ok := record.(*bsky.GraphFollow)
	if !ok {
		return false
	}
	return fi.IsFurryFeedDID(follow.Subject)
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
	record typegen.CBORMarshaler,
) (err error) {
	ctx, span := tracer.Start(ctx, "firehose_ingester.handle_record_create")
	defer func() {
		endSpan(span, err)
	}()
	span.SetAttributes(recordUriAttr(recordUri))

	actor := fi.actorCache.GetByDID(repoDID)
	if actor == nil {
		feedFollow := fi.isFurryFeedFollow(record)
		// If it's an unknown actor, and they've interacted, add em to
		// the candidate actor store with pending status. Otherwise, ignore
		// them.
		if !(feedFollow) {
			return nil
		}
		fi.log.Info(
			"unknown actor interacted, adding to db as pending",
			zap.String("did", repoDID),
			zap.Bool("feed_follow", feedFollow),
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

	switch data := record.(type) {
	case *bsky.FeedPost:
		err := fi.handleFeedPostCreate(ctx, repoDID, recordUri, data)
		if err != nil {
			return fmt.Errorf("handling app.bsky.feed.post create: %w", err)
		}
	case *bsky.FeedLike:
		err := fi.handleFeedLikeCreate(ctx, repoDID, recordUri, data)
		if err != nil {
			return fmt.Errorf("handling app.bsky.feed.like: %w", err)
		}
	case *bsky.GraphFollow:
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
		err = fi.handleGraphFollowDelete(ctx, repoDID, recordUri)
	default:
		span.AddEvent("ignoring record due to unrecognized type")
	}

	return
}

func (fi *FirehoseIngester) handleRecordUpdate(
	ctx context.Context,
	repoDID string,
	commitCID cid.Cid,
	recordUri string,
	updatedAt time.Time,
	record typegen.CBORMarshaler,
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

	switch data := record.(type) {
	case *bsky.ActorProfile:
		err := fi.handleActorProfileUpdate(ctx, repoDID, commitCID, recordUri, updatedAt, data)
		if err != nil {
			return fmt.Errorf("handling app.bsky.actor.profile update: %w", err)
		}
	default:
		span.AddEvent("ignoring record due to unrecognized type")
	}

	return nil
}
