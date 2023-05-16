package feedserver

import (
	"encoding/json"
	"fmt"
	"github.com/strideynet/bsky-furry-feed/store"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

var pinnedPost = ""

type getFeedSkeletonParameters struct {
	cursor string
	limit  int
	feed   string
}

func getFeedSkeletonHandler(
	log *zap.Logger, queries *store.Queries,
) (string, http.Handler) {
	var h http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
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
				handleErr(w, log, err)
				return
			}
			params.limit = limit
			if limit < 1 {
				handleErr(w, log, fmt.Errorf("limit too low (%d)", limit))
				return
			}
			if limit > 100 {
				handleErr(w, log, fmt.Errorf("limit too high (%d)", limit))
				return
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
			// Inject a pinned post at the top of the first page. We can use
			// this for service outage notifications etc.
			if pinnedPost != "" {
				output.Feed = append(output.Feed, post{
					Post: pinnedPost,
				})
			}

			posts, err := queries.ListCandidatePostsForFeed(
				r.Context(),
				int32(params.limit),
			)
			if err != nil {
				handleErr(
					w,
					log,
					fmt.Errorf("fetching candidate posts: %w", err),
				)
				return
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
		_ = encoder.Encode(output)
	}
	return "/xrpc/app.bsky.feed.getFeedSkeleton", otelhttp.NewHandler(h, "get_feed_skeleton")
}
