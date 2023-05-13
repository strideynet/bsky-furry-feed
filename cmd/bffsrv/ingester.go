package main

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
	bff "github.com/strideynet/bsky-furry-feed"
	"github.com/strideynet/bsky-furry-feed/store"
	typegen "github.com/whyrusleeping/cbor-gen"
	"go.uber.org/zap"
	"net/http"
	"sync"
	"time"
)

type candidateRepositoryCache struct {
	store  *store.Queries
	cached map[string]bff.CandidateRepository
	mu     sync.RWMutex
	log    *zap.Logger
}

func (crc *candidateRepositoryCache) GetByDID(
	did string,
) *bff.CandidateRepository {
	crc.mu.RLock()
	defer crc.mu.RUnlock()
	v, ok := crc.cached[did]
	if ok {
		return &v
	}
	return nil
}

func (crc *candidateRepositoryCache) fetch(ctx context.Context) error {
	crc.log.Info("starting cache fill")
	data, err := crc.store.ListCandidateRepositories(ctx)
	if err != nil {
		return fmt.Errorf("listing candidate repositories: %w", err)
	}

	mapped := map[string]bff.CandidateRepository{}
	for _, cr := range data {
		mapped[cr.DID] = bff.CandidateRepositoryFromStore(cr)
	}

	crc.mu.Lock()
	defer crc.mu.Unlock()
	crc.cached = mapped
	crc.log.Info("finished cache fill", zap.Int("count", len(mapped)))
	return nil
}

const workerCount = 3

type FirehoseIngester struct {
	stop  chan struct{}
	log   *zap.Logger
	crc   *candidateRepositoryCache
	store *store.Queries
}

func (fi *FirehoseIngester) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	subscribeUrl := "wss://bsky.social/xrpc/com.atproto.sync.subscribeRepos"

	con, _, err := websocket.DefaultDialer.Dial(subscribeUrl, http.Header{})
	if err != nil {
		return fmt.Errorf("dialing websocket: %w", err)
	}

	go func() {
		<-fi.stop
		fi.log.Info("closing websocket connection")
		if err := con.Close(); err != nil {
			fi.log.Error(
				"error occurred closing websocket",
				zap.Error(err),
			)
			return
		}
		cancel()
		fi.log.Info("closed websocket")
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
					// 30 seconds max to deal with any batch. This prevents a worker
					// hanging.
					ctx, cancel := context.WithTimeout(ctx, time.Second*30)
					defer cancel()
					if err := fi.handleCommit(ctx, evt); err != nil {
						fi.log.Error("failed to handle repo commit")
					}
				}
			}
		}()
	}

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
	fi.log.Info("waiting for workers to finish")
	workerWg.Wait()
	fi.log.Info("workers finished")

	return err
}

func (fi *FirehoseIngester) Stop() {
	fi.log.Info("stopping firehose ingester")
	// TODO: Tidier shutdown of websocket/workers order
	close(fi.stop)
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
	rr, err := repo.ReadRepoFromCar(ctx, bytes.NewReader(evt.Blocks))
	if err != nil {
		return fmt.Errorf("reading repo from car %w", err)
	}
	for _, op := range evt.Ops {
		// Ignore anything that isn't a new record being added
		if repomgr.EventKind(op.Action) != repomgr.EvtKindCreateRecord {
			continue
		}

		uri := fmt.Sprintf("at://%s/%s", evt.Repo, op.Path)
		log := fi.log.With(
			zap.String("repository", evt.Repo),
			zap.String("repository.comment", candidateUser.Comment),
			zap.String("action", op.Action),
			zap.String("uri", uri),
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
		if err := fi.handleRecordCreate(ctx, log, evt.Repo, uri, record); err != nil {
			return fmt.Errorf("handleRecordCreate: %w", err)
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

	// TODO: Manage goroutine worker pool for handling records

	switch data := record.(type) {
	case *bsky.FeedPost:
		if data.Reply == nil {
			createdAt, err := time.Parse("2006-01-02T15:04:05.000Z", data.CreatedAt)
			if err != nil {
				return fmt.Errorf("parsing post time: %w", err)
			}
			err = fi.store.CreateCandidatePost(
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
				return fmt.Errorf("creating candidate post")
			}
		} else {
			log.Info("ignoring reply")
		}
	default:
		log.Info("ignoring unhandled record type")
	}

	return nil
}
