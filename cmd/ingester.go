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
)

type RecordHandler struct {
	log *slog.Logger
}

func (rh *RecordHandler) HandleCreate(
	repoDID string,
	recordPath string,
	record typegen.CBORMarshaler,
) {
	log := rh.log.With("repo", repoDID, "record_path", recordPath)
	if repoDID != noahDid {
		return
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
}

type WebSocketFirehose struct {
	stop                chan struct{}
	log                 *slog.Logger
	RecordCreateHandler func(repoDID string, recordPath string, r typegen.CBORMarshaler)
}

func (fh *WebSocketFirehose) Start() error {
	ctx := context.Background()
	subscribeUrl := "wss://bsky.social/xrpc/com.atproto.sync.subscribeRepos"

	con, _, err := websocket.DefaultDialer.Dial(subscribeUrl, http.Header{})
	if err != nil {
		return fmt.Errorf("dialing websocket: %w", err)
	}

	go func() {
		<-fh.stop
		fh.log.Info("Closing websocket connection.")
		con.Close()
	}()

	return events.HandleRepoStream(ctx, con, &events.RepoStreamCallbacks{
		RepoCommit: func(evt *atproto.SyncSubscribeRepos_Commit) error {
			ctx, span := tracer.Start(ctx, "WebSocketFirehose.HandleRepoCommit")
			defer span.End()

			log := fh.log.With("repo", evt.Repo)
			log.Debug("Commit event received", "opsCount", len(evt.Ops))
			rr, err := repo.ReadRepoFromCar(ctx, bytes.NewReader(evt.Blocks))
			if err != nil {
				return fmt.Errorf("reading repo from car %w", err)
			}
			for _, op := range evt.Ops {
				log := log.With("path", op.Path)
				// Ignore anything that isn't a new record being added
				if repomgr.EventKind(op.Action) != repomgr.EvtKindCreateRecord {
					log.Debug("Ignoring op", "action", op.Action)
					continue
				}
				recordCid, record, err := rr.GetRecord(ctx, op.Path)
				if err != nil {
					if errors.Is(err, lexutil.ErrUnrecognizedType) {
						continue
					}
					return fmt.Errorf("getting record for op: %w", err)
				}
				if lexutil.LexLink(recordCid) != *op.Cid {
					return fmt.Errorf("mismatch in record and op cid: %s != %s", recordCid, *op.Cid)
				}
				log.Debug("Record fetched", "record", record, "type", fmt.Sprintf("%T", record))
				fh.RecordCreateHandler(evt.Repo, op.Path, record)
			}

			return nil
		},
	})
}

func (fh *WebSocketFirehose) Stop() {
	fh.log.Info("Stopping WebSocket Firehose.")
	close(fh.stop)
}

const noahDid = "did:plc:dllwm3fafh66ktjofzxhylwk"
