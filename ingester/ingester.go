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
	"github.com/strideynet/bsky-furry-feed/bluesky"
	"github.com/strideynet/bsky-furry-feed/store"
	typegen "github.com/whyrusleeping/cbor-gen"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
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
	crc     *CandidateRepositoryCache
	queries *store.Queries

	// configuration
	subscribeURL    string
	workerCount     int
	workItemTimeout time.Duration
}

func NewFirehoseIngester(
	log *zap.Logger, queries *store.Queries, crc *CandidateRepositoryCache,
) *FirehoseIngester {
	return &FirehoseIngester{
		log:     log,
		crc:     crc,
		queries: queries,

		subscribeURL:    "wss://bsky.social/xrpc/com.atproto.sync.subscribeRepos",
		workerCount:     5,
		workItemTimeout: time.Second * 30,
	}
}

func (fi *FirehoseIngester) Start(ctx context.Context) error {
	eg, egCtx := errgroup.WithContext(ctx)

	// Unbuffered channel so that the websocket will stop reading if the workers
	// are not ready. In future, we may want to consider some reasonable
	// buffering to account for short spikes in event rates.
	evtChan := make(chan *atproto.SyncSubscribeRepos_Commit)
	eg.Go(func() error {
		workerWg := sync.WaitGroup{}
		for n := 1; n < fi.workerCount; n++ {
			workerWg.Add(1)
			go func() {
				defer workerWg.Done()
				for {
					select {
					case <-ctx.Done():
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
		ctx, cancel := context.WithCancel(egCtx)
		defer cancel()

		con, _, err := websocket.DefaultDialer.DialContext(
			ctx, fi.subscribeURL, http.Header{},
		)
		if err != nil {
			return fmt.Errorf("dialing websocket: %w", err)
		}

		go func() {
			<-ctx.Done()
			fi.log.Info("closing websocket subscription")
			if err := con.Close(); err != nil {
				fi.log.Error(
					"error occurred closing websocket",
					zap.Error(err),
				)
			}
			fi.log.Info("closed websocket subscription")
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
	// Dispose of events from non-candidate repositories
	candidateUser := fi.crc.GetByDID(evt.Repo)
	if candidateUser == nil {
		return nil
	}
	// TODO: Find a way to use tail-based sampling so that we can capture this trace
	// before candidateUser is run and ensure we always capture candidateUser
	// traces.
	ctx, span := tracer.Start(ctx, "firehose_ingester.handle_commit")
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
	case *bsky.GraphFollow:
		err := fi.handleGraphFollowCreate(ctx, log, repoDID, recordUri, data)
		if err != nil {
			return fmt.Errorf("handling app.bsky.graph.follow: %w", err)
		}
	default:
		log.Info("ignoring record create due to handled type")
	}

	return nil
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
		createdAt, err := bluesky.ParseTime(data.CreatedAt)
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

	createdAt, err := bluesky.ParseTime(data.CreatedAt)
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

var discordWebhookGraphFollow = os.Getenv("DISCORD_WEBHOOK_GRAPH_FOLLOW")

func (fi *FirehoseIngester) handleGraphFollowCreate(
	ctx context.Context,
	log *zap.Logger,
	repoDID string,
	recordUri string,
	data *bsky.GraphFollow,
) error {
	ctx, span := tracer.Start(ctx, "firehose_ingester.handle_graph_follow_create")
	defer span.End()

	if fi.crc.GetByDID(data.Subject) != nil {
		// We aren't interested in repositories we already track.
		return nil
	}
	if discordWebhookGraphFollow == "" {
		log.Warn("no webhook configured for graph follow")
		return nil
	}

	var jsonStr = []byte(fmt.Sprintf(`
{
    "username": "bff-system",
    "content": "**Furry follow?**\n**Source repository:** %s \n**Followed:** https://psky.app/profile/%s"
}`, repoDID, data.Subject))
	req, err := http.NewRequest("POST", discordWebhookGraphFollow, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("sending discord webhook: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
