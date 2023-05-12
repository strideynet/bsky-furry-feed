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
	stop chan struct{}
	log  *zap.Logger
	crc  *candidateRepositoryCache
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
					if err := fi.handleRepoCommit(ctx, evt); err != nil {
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

func (fi *FirehoseIngester) handleRepoCommit(rootCtx context.Context, evt *atproto.SyncSubscribeRepos_Commit) error {
	// Dispose of events from non-candidate repositories
	candidateUser := fi.crc.GetByDID(evt.Repo)
	if candidateUser == nil {
		return nil
	}
	// TODO: Find a way to use tail-based sampling so that we can capture this trace
	// before candidateUser is run and ensure we always capture candidateUser
	// traces.
	ctx, span := tracer.Start(rootCtx, "FirehoseIngester.handleRepoCommit")
	defer span.End()
	log := fi.log.With(
		zap.String("candidate_repository", evt.Repo),
		zap.String("candidate_repository.comment", candidateUser.Comment),
	)
	rr, err := repo.ReadRepoFromCar(ctx, bytes.NewReader(evt.Blocks))
	if err != nil {
		return fmt.Errorf("reading repo from car %w", err)
	}
	for _, op := range evt.Ops {
		log := log.With(
			zap.String("path", op.Path),
			zap.String("action", op.Action),
		)
		// Ignore anything that isn't a new record being added
		if repomgr.EventKind(op.Action) != repomgr.EvtKindCreateRecord {
			log.Debug("ignoring op due to action")
			continue
		}
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
		if err := fi.handleRecordCreate(ctx, log, evt.Repo, op.Path, record); err != nil {
			return fmt.Errorf("handleRecordCreate: %w", err)
		}
	}

	return nil
}

func (fi *FirehoseIngester) handleRecordCreate(
	ctx context.Context,
	log *zap.Logger,
	repoDID string,
	recordPath string,
	record typegen.CBORMarshaler,
) error {
	ctx, span := tracer.Start(ctx, "FirehoseIngester.handleRecordCreate")
	defer span.End()
	log.Info("handling record create")

	// TODO: Manage goroutine worker pool for handling records

	switch data := record.(type) {
	case *bsky.FeedLike:
		log.Info(
			"like",
			zap.Any("data", data),
			zap.Any("subject", data.Subject),
		)
	case *bsky.FeedPost:
		postType := "post"
		if data.Reply != nil {
			postType = "reply"
		}
		log.Info(postType, zap.Any("data", data))
	case *bsky.FeedRepost:
		log.Info("repost", zap.Any("data", data))
	case *bsky.GraphFollow:
		log.Info("follow", zap.Any("data", data))
	default:
		log.Info(
			"unhandled record type",
			zap.String("type", fmt.Sprintf("%T", data)),
		)
	}

	return nil
}
