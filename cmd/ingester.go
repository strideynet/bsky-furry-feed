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

const noahDid = "did:plc:dllwm3fafh66ktjofzxhylwk"

type FirehoseIngester struct {
	stop chan struct{}
	log  *slog.Logger
}

func (fi *FirehoseIngester) Start() error {
	ctx := context.Background()
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
		fi.log.Info("closed websocket")

	}()

	workerWg := sync.WaitGroup{}
	err = events.HandleRepoStream(ctx, con, &events.RepoStreamCallbacks{
		RepoCommit: func(evt *atproto.SyncSubscribeRepos_Commit) error {
			// TODO: Make this a worker pool limited by size.
			workerWg.Add(1)
			go func() {
				defer workerWg.Done()
				ctx, cancel := context.WithTimeout(ctx, time.Second*30)
				defer cancel()
				if err := fi.handleRepoCommit(ctx, evt); err != nil {
					fi.log.Error("failed to handle repo commit")
				}
			}()

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
	if repoDID != noahDid {
		return nil
	}
	log.Info("Handling record")

	// TODO: Manage goroutine worker pool for handling records

	switch data := record.(type) {
	case *bsky.FeedLike:
		log.Info("Like", "data", data, "subject", data.Subject)
	case *bsky.FeedPost:
		postType := "Post"
		if data.Reply != nil {
			postType = "Reply"
		}
		log.Info(postType, "data", data)
	default:
		log.Info("Unhandled record type", "type", fmt.Sprintf("%T", data))
	}

	return nil
}
