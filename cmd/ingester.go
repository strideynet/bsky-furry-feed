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
	typegen "github.com/whyrusleeping/cbor-gen"
	"golang.org/x/exp/slog"
	"net/http"
	"sync"
	"time"
)

const workerCount = 3

type FirehoseIngester struct {
	stop        chan struct{}
	log         *slog.Logger
	usersGetter StaticCandidateUsers
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
			fi.log.Error("error occurred closing websocket", "err", err)
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
	ctx, span := tracer.Start(rootCtx, "FirehoseIngester.handleRepoCommit")
	defer span.End()
	log := fi.log.With("repo", evt.Repo)

	// Only track events from opted-in "furries"
	candidateUser := fi.usersGetter.GetByDID(evt.Repo)
	if candidateUser == nil {
		return nil
	}
	log = fi.log.With("candidateUser", candidateUser.comment)

	log.Debug("commit event received", "opsCount", len(evt.Ops))
	rr, err := repo.ReadRepoFromCar(ctx, bytes.NewReader(evt.Blocks))
	if err != nil {
		return fmt.Errorf("reading repo from car %w", err)
	}
	for _, op := range evt.Ops {
		log := log.With("path", op.Path, "action", op.Action)
		// Ignore anything that isn't a new record being added
		if repomgr.EventKind(op.Action) != repomgr.EvtKindCreateRecord {
			log.Debug("ignoring op", "action", op.Action)
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
		log.Debug("record fetched", "record", record, "type", fmt.Sprintf("%T", record))
		if err := fi.handleRecordCreate(ctx, log, evt.Repo, op.Path, record); err != nil {
			return fmt.Errorf("handleRecordCreate: %w", err)
		}
	}

	return nil
}

func (fi *FirehoseIngester) handleRecordCreate(
	ctx context.Context,
	log *slog.Logger,
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
		log.Info("like", "data", data, "subject", data.Subject)
	case *bsky.FeedPost:
		postType := "post"
		if data.Reply != nil {
			postType = "reply"
		}
		log.Info(postType, "data", data)
	case *bsky.FeedRepost:
		log.Info("repost", "data", data)
	case *bsky.GraphFollow:
		log.Info("follow", "data", data)
	default:
		log.Info("unhandled record type", "type", fmt.Sprintf("%T", data))
	}

	return nil
}
