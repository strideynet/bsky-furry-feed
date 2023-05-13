package feedserver

import (
	"encoding/json"
	"github.com/strideynet/bsky-furry-feed/store"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

func getFeedSkeletonHandler(
	log *zap.Logger, queries *store.Queries,
) (string, http.HandlerFunc) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

			posts, err := queries.ListCandidatePostsForFeed(r.Context(), int32(params.limit))
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
	})
	return "/xrpc/app.bsky.feed.getFeedSkeleton", h
}
