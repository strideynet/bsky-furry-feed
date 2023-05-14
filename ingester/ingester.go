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
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/strideynet/bsky-furry-feed/store"
	typegen "github.com/whyrusleeping/cbor-gen"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"net/http"
	"sync"
	"time"
)

var tracer = otel.Tracer("github.com/strideynet/bsky-furry-feed/ingester")

var eventsProcessed = promauto.NewSummaryVec(prometheus.SummaryOpts{
	Name: "bff_ingester_firehose_event_duration_seconds",
	Help: "The total number of events from the firehose processed by workers",
}, []string{"type"})

const workerCount = 3

type FirehoseIngester struct {
	log     *zap.Logger
	crc     *CandidateRepositoryCache
	queries *store.Queries
}

func NewFirehoseIngester(log *zap.Logger, queries *store.Queries, crc *CandidateRepositoryCache) *FirehoseIngester {
	return &FirehoseIngester{
		log:     log,
		crc:     crc,
		queries: queries,
	}
}

func (fi *FirehoseIngester) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	subscribeUrl := "wss://bsky.social/xrpc/com.atproto.sync.subscribeRepos"

	con, _, err := websocket.DefaultDialer.Dial(subscribeUrl, http.Header{})
	if err != nil {
		return fmt.Errorf("dialing websocket: %w", err)
	}

	go func() {
		<-ctx.Done()
		fi.log.Info("stopping firehose ingester")
		fi.log.Info("closing websocket connection")
		if err := con.Close(); err != nil {
			fi.log.Error(
				"error occurred closing websocket",
				zap.Error(err),
			)
		}
		fi.log.Info("closed websocket")
		cancel()
	}()

	// Unbuffered channel so that the websocket will stop reading if the workers
	// are not redy.
	evtChan := make(chan *atproto.SyncSubscribeRepos_Commit)
	workerWg := sync.WaitGroup{}
	for workerN := 1; workerN < workerCount; workerN++ {
		workerWg.Add(1)
		go func() {
			defer workerWg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case evt := <-evtChan:
					start := time.Now()
					// 30 seconds max to deal with any batch. This prevents a worker
					// hanging.
					ctx, cancel := context.WithTimeout(ctx, time.Second*30)
					if err := fi.handleCommit(ctx, evt); err != nil {
						fi.log.Error("failed to handle repo commit", zap.Error(err))
					}
					eventsProcessed.WithLabelValues("repo_commit").Observe(time.Since(start).Seconds())
					cancel()
				}
			}
		}()
	}
	// TODO: sometimes stream exits of own accord, lets recover from that.
	err = events.HandleRepoStream(ctx, con, &events.RepoStreamCallbacks{
		RepoCommit: func(evt *atproto.SyncSubscribeRepos_Commit) error {
			select {
			case <-ctx.Done():
				// Ensure we don't get stuck waiting for a worker even if the
				// server has shutdown.
				return nil
			case evtChan <- evt:
			}
			return nil
		},
	})
	if err != nil {
		return err
	}
	fi.log.Info("waiting for workers to finish")
	workerWg.Wait()
	fi.log.Info("workers finished")

	return err
}

func (fi *FirehoseIngester) handleCommit(rootCtx context.Context, evt *atproto.SyncSubscribeRepos_Commit) error {
	// Dispose of events from non-candidate repositories
	candidateUser := fi.crc.GetByDID(evt.Repo)
	if candidateUser == nil {
		return nil
	}
	// TODO: Find a way to use tail-based sampling so that we can capture this trace
	// before candidateUser is run and ensure we always capture candidateUser
	// traces.
	ctx, span := tracer.Start(rootCtx, "firehose_ingester.handle_commit")
	defer span.End()
	span.SetAttributes(
		attribute.String("candidate_repository.did", evt.Repo),
	)
	log := fi.log.With(zap.String("candidate_repository.did", evt.Repo))
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
		if err := fi.handleRecordCreate(ctx, log, evt.Repo, uri, record); err != nil {
			return fmt.Errorf("handling record create: %w", err)
		}
	}

	return nil
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
	log.Info("handling record create", zap.Any("record", record))

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
	default:
		log.Info("ignoring record create due to handled type")
	}

	return nil
}

func parseTime(str string) (time.Time, error) {
	t, err := time.Parse("2006-01-02T15:04:05.999999999Z", str)
	if err != nil {
		return time.Time{}, fmt.Errorf("parsing post time: %w", err)
	}
	return t, nil
}

func (fi *FirehoseIngester) handleFeedPostCreate(
	ctx context.Context,
	log *zap.Logger,
	repoDID string,
	recordUri string,
	data *bsky.FeedPost,
) error {
	ctx, span := tracer.Start(ctx, "firehose_ingester.handle_feed_post_create")
	defer span.End()
	if data.Reply == nil {
		createdAt, err := parseTime(data.CreatedAt)
		if err != nil {
			return fmt.Errorf("parsing post time: %w", err)
		}
		err = fi.queries.CreateCandidatePost(
			ctx,
			store.CreateCandidatePostParams{
				URI:           recordUri,
				RepositoryDID: repoDID,
				CreatedAt: pgtype.Timestamptz{
					Time:  createdAt,
					Valid: true,
				},
				IndexedAt: pgtype.Timestamptz{
					Time:  time.Now(),
					Valid: true,
				},
			},
		)
		if err != nil {
			return fmt.Errorf("creating candidate post: %w", err)
		}
	} else {
		log.Info("ignoring reply")
	}
	return nil
}

func (fi *FirehoseIngester) handleFeedLikeCreate(
	ctx context.Context,
	log *zap.Logger,
	repoDID string,
	recordUri string,
	data *bsky.FeedLike,
) error {
	ctx, span := tracer.Start(ctx, "firehose_ingester.handle_feed_like_create")
	defer span.End()

	createdAt, err := parseTime(data.CreatedAt)
	if err != nil {
		return fmt.Errorf("parsing like time: %w", err)
	}
	err = fi.queries.CreateCandidateLike(
		ctx,
		store.CreateCandidateLikeParams{
			URI:           recordUri,
			RepositoryDID: repoDID,
			SubjectUri:    data.Subject.Uri,
			CreatedAt: pgtype.Timestamptz{
				Time:  createdAt,
				Valid: true,
			},
			IndexedAt: pgtype.Timestamptz{
				Time:  time.Now(),
				Valid: true,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("creating candidate like: %w", err)
	}

	return nil
}
