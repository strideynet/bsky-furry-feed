package main

import (
	"encoding/json"
	"fmt"
	"github.com/strideynet/bsky-furry-feed/store"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type getFeedSkeletonParameters struct {
	cursor string
	limit  int
	feed   string
}

const hostname = "feed.ottr.sh"

// https://feed-generator.skyfeed.app/xrpc/app.bsky.feed.getFeedSkeleton?feed=did:web:feed-generator.skyfeed.app/app.bsky.feed.generator/posts-with-links
func feedServer(log *zap.Logger, st *store.Queries) *http.Server {
	mux := &http.ServeMux{}
	mux.HandleFunc("/.well-known/did.json", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(fmt.Sprintf(`{"@context":["https://www.w3.org/ns/did/v1"],"id":"did:web:%s","service":[{"id":"#bsky_fg","type":"BskyFeedGenerator","serviceEndpoint":"https://%s"}]}`, hostname, hostname)))
	})
	mux.Handle("/xrpc/app.bsky.feed.getFeedSkeleton", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		params := getFeedSkeletonParameters{
			cursor: q.Get("cursor"),
			feed:   q.Get("feed"),
			limit:  50,
		}
		limitStr := q.Get("limit")
		if limitStr != "" {
			limit, err := strconv.Atoi(limitStr)
			if err != nil {
				panic(err)
			}
			params.limit = limit
			if limit < 1 {
				panic("limit too low")
			}
			if limit > 100 {
				panic("limit too high")
			}
		}
		log.Info(
			"get feed skeleton request",
			zap.String("feed", params.feed),
			zap.String("cursor", params.cursor),
			zap.Int("limit", params.limit),
		)

		type post struct {
			Post string `json:"post"`
		}
		output := struct {
			Cursor string `json:"cursor"`
			Feed   []post `json:"feed"`
		}{
			Feed: []post{},
		}

		if params.cursor == "" {
			// Inject a pinned skeet at the top.
			pinnedPost := "at://did:plc:dllwm3fafh66ktjofzxhylwk/app.bsky.feed.post/3jvmbtpvjlq2j"
			output.Feed = append(output.Feed, post{
				Post: pinnedPost,
			})

			posts, err := st.ListCandidatePostsForFeed(r.Context(), int32(params.limit))
			if err != nil {
				log.Error("failed to fetch posts", zap.Error(err))
				panic(err)
			}

			for _, p := range posts {
				// Remove pinned skeet to avoid showing it twice
				if p.URI == pinnedPost {
					continue
				}
				output.Feed = append(output.Feed, post{
					Post: p.URI,
				})
			}
			// TODO: index_at timestamp of last post
			output.Cursor = "end"
		} else {
			output.Cursor = ""
		}

		w.WriteHeader(200)
		encoder := json.NewEncoder(w)
		encoder.Encode(output)
		// TODO: Handle err.
	}))
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("request", zap.Any("r", r))
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("boo!"))
	}))

	return &http.Server{
		Addr:    ":1337",
		Handler: mux,
	}
}
