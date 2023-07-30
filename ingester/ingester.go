package ingester

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/bluesky-social/indigo/util"

	"net/http"
	"sync"
	"time"

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

type FirehoseIngester struct {
	// dependencies
	log   *zap.Logger
	crc   *ActorCache
	store *store.PGXStore

	// configuration
	subscribeURL    string
	workerCount     int
	workItemTimeout time.Duration
}

func NewFirehoseIngester(
	log *zap.Logger, store *store.PGXStore, crc *ActorCache, pdsHost string,
) *FirehoseIngester {
	return &FirehoseIngester{
		log:   log,
		crc:   crc,
		store: store,

		subscribeURL:    pdsHost + "/xrpc/com.atproto.sync.subscribeRepos",
		workerCount:     8,
		workItemTimeout: time.Second * 30,
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
		return events.HandleRepoStream(ctx, con, &events.SequentialScheduler{
			Do: callbacks.EventHandler,
		})
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
				return fmt.Errorf("getting record for op: %w", err)
			}

			// Ensure there isn't a mismatch between the reference and the found
			// object.
			if lexutil.LexLink(recordCid) != *op.Cid {
				return fmt.Errorf("mismatch in record and op cid: %s != %s", recordCid, *op.Cid)
			}

			if err := fi.handleRecordCreate(
				ctx, evt.Repo, uri, record,
			); err != nil {
				return fmt.Errorf("handling record create: %w", err)
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

func (fi *FirehoseIngester) isFurryFeedFollow(record typegen.CBORMarshaler) bool {
	follow, ok := record.(*bsky.GraphFollow)
	if !ok {
		return false
	}

	// TODO: Make this not hard coded
	// https://bsky.app/profile/furryli.st
	return follow.Subject == "did:plc:jdkvwye2lf4mingzk7qdebzc"
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

	actor := fi.crc.GetByDID(repoDID)
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
		if err := fi.crc.CreatePendingCandidateActor(ctx, repoDID); err != nil {
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

	actor := fi.crc.GetByDID(repoDID)
	if actor == nil {
		// if we don’t know the actor, we don’t have their data
		return
	}

	parsedUri, err := util.ParseAtUri(recordUri)
	if err != nil {
		return fmt.Errorf("parsing uri: %w", err)
	}

	switch parsedUri.Collection {
	case "app.bsky.feed.post":
		err = fi.handleFeedPostDelete(ctx, recordUri)
	case "app.bsky.feed.like":
		err = fi.handleFeedLikeDelete(ctx, recordUri)
	case "app.bsky.graph.follow":
		err = fi.handleFeedFollowDelete(ctx, recordUri)
	default:
		span.AddEvent("ignoring record due to unrecognized type")
	}

	return
}
