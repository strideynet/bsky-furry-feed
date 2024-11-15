package ingester

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync/atomic"

	"time"

	"github.com/bluesky-social/indigo/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	jsclient "github.com/bluesky-social/jetstream/pkg/client"
	jsparallel "github.com/bluesky-social/jetstream/pkg/client/schedulers/parallel"
	"github.com/bluesky-social/jetstream/pkg/models"

	v1 "github.com/strideynet/bsky-furry-feed/proto/bff/v1"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/bluesky-social/indigo/api/bsky"
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

var flushedWorkerCursor = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "bff_ingester_flushed_worker_cursor",
	Help: "The current cursor flushed to persistent storage.",
})

type actorCacher interface {
	GetByDID(did string) *v1.Actor
	CreatePendingCandidateActor(ctx context.Context, did string) (err error)
}

type FirehoseIngester struct {
	// dependencies
	log        *zap.Logger
	actorCache actorCacher
	store      *store.PGXStore

	// configuration
	jetstreamURL        string
	workerCount         int
	workItemTimeout     time.Duration
	cursorFlushInterval time.Duration
}

func NewFirehoseIngester(
	log *zap.Logger, store *store.PGXStore, crc *ActorCache,
) *FirehoseIngester {
	return &FirehoseIngester{
		log:        log,
		actorCache: crc,
		store:      store,

		jetstreamURL:        "wss://jetstream2.us-east.bsky.network/subscribe",
		workerCount:         20,
		workItemTimeout:     time.Second * 30,
		cursorFlushInterval: time.Second * 10,
	}
}

func (fi *FirehoseIngester) Start(ctx context.Context) (err error) {
	eg, ctx := errgroup.WithContext(ctx)
	slogLog := slog.Default() // TODO: Switch from Zap to Slog

	jsCfg := jsclient.DefaultClientConfig()
	jsCfg.WebsocketURL = fi.jetstreamURL
	jsCfg.WantedCollections = []string{
		"app.bsky.actor.profile",
		"app.bsky.feed.like",
		"app.bsky.feed.post",
		"app.bsky.graph.follow",
	}

	var activeCursor atomic.Int64
	sched := jsparallel.NewScheduler(
		fi.workerCount,
		"jetstream",
		slogLog,
		func(ctx context.Context, e *models.Event) error {
			start := time.Now()
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

			activeCursor.Store(e.TimeUS)
			workItemsProcessed.
				WithLabelValues("repo_commit").
				Observe(time.Since(start).Seconds())
			return nil
		},
	)

	jsClient, err := jsclient.NewClient(
		jsCfg, slogLog, sched,
	)
	if err != nil {
		return fmt.Errorf("creating jetstream client: %w", err)
	}

	initCursor, err := fi.store.GetJetstreamCursor(ctx)
	if err != nil {
		return fmt.Errorf("get jetstream cursor: %w", err)
	}
	if initCursor == -1 {
		initCursor = time.Now().UnixMicro()
	}
	// Step back a few minutes to allow recovery
	initCursor = time.UnixMicro(initCursor).Add(-1 * time.Minute).UnixMicro()

	eg.Go(func() error {
		for {
			cursor := activeCursor.Load()
			if cursor == 0 {
				cursor = initCursor
			}

			fi.log.Info(
				"starting ingestion",
				zap.Int64("cursor", cursor),
				zap.String("cursor_time", time.UnixMicro(cursor).String()),
			)
			if err := jsClient.ConnectAndRead(ctx, &cursor); err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}
				fi.log.Error("jetstream client encountered an error, restarting", zap.Error(err))
			}
		}
	})

	flushCursor := func(ctx context.Context) {
		cursor := activeCursor.Load()
		if cursor == 0 {
			fi.log.Warn("no cursor value to persist")
			return
		}
		if cursor <= initCursor {
			// attempt to avoid a scenario where a crash loop sends the
			// cursor further and further into the past.
			fi.log.Warn("not setting cursor to avoid regression")
			return
		}
		if err := fi.store.SetJetstreamCursor(ctx, cursor); err != nil {
			fi.log.Warn("failed to flush cursor", zap.Error(err))
			return
		}

		fi.log.Info(
			"successfully flushed cursor",
			zap.Int64("cursor", cursor),
			zap.String("cursor_time", time.UnixMicro(cursor).String()),
		)
		flushedWorkerCursor.Set(float64(cursor))
	}

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

			flushCursor(ctx)
		}
	})

	err = eg.Wait()

	// Perform a final cursor flush.
	fi.log.Info("performing final flush of cursor")
	exitCtx, cancelExitCtx := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelExitCtx()
	flushCursor(exitCtx)

	return err
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
			return fmt.Errorf("unmarshalling app.bsky.graph.follow: %w", err)
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
			return fmt.Errorf("unmarshalling app.bsky.feed.post: %w", err)
		}
		err := fi.handleFeedPostCreate(ctx, repoDID, recordUri, data)
		if err != nil {
			return fmt.Errorf("handling app.bsky.feed.post create: %w", err)
		}
	case "app.bsky.feed.like":
		data := &bsky.FeedLike{}
		if err := json.Unmarshal(record, data); err != nil {
			return fmt.Errorf("unmarshalling app.bsky.feed.like: %w", err)
		}
		err := fi.handleFeedLikeCreate(ctx, repoDID, recordUri, data)
		if err != nil {
			return fmt.Errorf("handling app.bsky.feed.like: %w", err)
		}
	case "app.bsky.graph.follow":
		data := &bsky.GraphFollow{}
		if err := json.Unmarshal(record, data); err != nil {
			return fmt.Errorf("unmarshalling app.bsky.graph.follow: %w", err)
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
			return fmt.Errorf("unmarshalling app.bsky.actor.profile: %w", err)
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
