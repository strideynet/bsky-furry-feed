package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/events"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/repo"
	"github.com/bluesky-social/indigo/repomgr"
	"github.com/gorilla/websocket"
	"github.com/oklog/run"
	"go.opentelemetry.io/otel"
	"golang.org/x/exp/slog"
	"net/http"
	"os"
	"strconv"
)

var tracer = otel.Tracer("github.com/strideynet/bsky-furry-feed")

func main() {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	err := runE(log)
	if err != nil {
		panic(err)
	}
}

func runE(log *slog.Logger) error {
	ctx := context.Background()

	runGroup := run.Group{}
	fireHoseCtx, fireHoseCancel := context.WithCancel(ctx)
	runGroup.Add(func() error {
		return fireHose(fireHoseCtx, log)
	}, func(err error) {
		fireHoseCancel()
	})

	srv := feedServer(log)
	runGroup.Add(func() error {
		return srv.ListenAndServe()
	}, func(err error) {
		srv.Close()
	})

	return runGroup.Run()
}

const noahDid = "did:plc:dllwm3fafh66ktjofzxhylwk"

func fireHose(ctx context.Context, log *slog.Logger) error {
	subscribeUrl := "wss://bsky.social/xrpc/com.atproto.sync.subscribeRepos"

	con, _, err := websocket.DefaultDialer.Dial(subscribeUrl, http.Header{})
	if err != nil {
		return fmt.Errorf("dialing websocket: %w", err)
	}

	return events.HandleRepoStream(ctx, con, &events.RepoStreamCallbacks{
		RepoCommit: func(evt *atproto.SyncSubscribeRepos_Commit) error {
			ctx, span := tracer.Start(ctx, "FireHose.HandleRepoCommit")
			defer span.End()

			if evt.Repo != noahDid {
				return nil
			}
			log := log.With("repo", evt.Repo)
			log.Info("Commit event received", "ops_count", len(evt.Ops))
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
				recordCid, encodedRecord, err := rr.GetRecord(ctx, op.Path)
				if err != nil {
					if errors.Is(err, lexutil.ErrUnrecognizedType) {
						continue
					}
					return fmt.Errorf("getting record for op: %w", err)
				}
				if lexutil.LexLink(recordCid) != *op.Cid {
					return fmt.Errorf("mismatch in record and op cid: %s != %s", recordCid, *op.Cid)
				}
				log.Debug("Record fetched", "encoded_record", encodedRecord, "type", fmt.Sprintf("%T", encodedRecord))
				switch data := encodedRecord.(type) {
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

			return nil
		},
	})
}

type getFeedSkeletonParameters struct {
	cursor string
	limit  int
	feed   string
}

func feedServer(log *slog.Logger) *http.Server {
	mux := &http.ServeMux{}
	mux.Handle("/xrpc/app.bsky.feed.getFeedSkeleton", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		params := getFeedSkeletonParameters{
			cursor: q.Get("cursor"),
			feed:   q.Get("feed"),
		}
		limitStr := q.Get("limit")
		if limitStr != "" {
			limit, err := strconv.Atoi(limitStr)
			if err != nil {
				panic(err)
			}
			params.limit = limit
		}

		w.WriteHeader(200)
		output := map[string]any{
			"cursor": "my-cursor",
			"feed":   []any{},
		}
		encoder := json.NewEncoder(w)
		encoder.Encode(output)
	}))
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("request")
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("boo!"))
	}))

	return &http.Server{
		Addr:    ":1337",
		Handler: mux,
	}
}
