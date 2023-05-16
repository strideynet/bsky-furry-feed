package feedserver

import (
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/strideynet/bsky-furry-feed/bluesky"
	"github.com/strideynet/bsky-furry-feed/store"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"strconv"
)

type getFeedSkeletonParams struct {
	cursor string
	limit  int
	feed   string
}

func parseGetFeedSkeletonParams(u *url.URL) (*getFeedSkeletonParams, error) {
	q := u.Query()
	params := getFeedSkeletonParams{
		cursor: q.Get("cursor"),
		feed:   q.Get("feed"),
		limit:  50, // Default value
	}
	// TODO: Parse "feed" into name of feed and ignore the hostname
	limitStr := q.Get("limit")
	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert 'limit' param to integer: %w", err)
		}
		if limit < 1 {
			return nil, fmt.Errorf("limit too low (%d)", limit)
		}
		if limit > 100 {
			return nil, fmt.Errorf("limit too high (%d)", limit)
		}
		params.limit = limit
	}

	return &params, nil
}

type getFeedSkeletonResponsePost struct {
	Post string `json:"post"`
}

type getFeedSkeletonResponse struct {
	Cursor string                        `json:"cursor"`
	Feed   []getFeedSkeletonResponsePost `json:"feed"`
}

func getFeedSkeletonHandler(
	log *zap.Logger, queries *store.Queries,
) (string, http.Handler) {
	var h http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		params, err := parseGetFeedSkeletonParams(r.URL)
		if err != nil {
			handleErr(w, log, err)
			return
		}
		log.Info(
			"get feed skeleton request",
			zap.String("feed", params.feed),
			zap.String("cursor", params.cursor),
			zap.Int("limit", params.limit),
		)

		var posts []store.CandidatePost
		if params.cursor == "" {
			// TODO: Reintroduce pinned top-post at a later date. This should
			// only be injected if no cursor has been specified.
			posts, err = queries.ListCandidatePostsForFeed(
				r.Context(),
				int32(params.limit),
			)
			if err != nil {
				handleErr(
					w,
					log,
					fmt.Errorf("fetching candidate posts without cursor: %w", err),
				)
				return
			}
		} else {
			cursorTime, err := bluesky.ParseTime(params.cursor)
			if err != nil {
				handleErr(
					w,
					log,
					fmt.Errorf("parsing cursor time: %w", err),
				)
				return
			}
			posts, err = queries.ListCandidatePostsForFeedWithCursor(
				r.Context(),
				store.ListCandidatePostsForFeedWithCursorParams{
					Limit: int32(params.limit),
					CreatedAt: pgtype.Timestamptz{
						Valid: true,
						Time:  cursorTime,
					},
				},
			)
			if err != nil {
				handleErr(
					w,
					log,
					fmt.Errorf("fetching candidate posts with cursor: %w", err),
				)
				return
			}
		}

		// Convert the selected posts to the getFeedSkeleton format
		output := getFeedSkeletonResponse{
			Feed: []getFeedSkeletonResponsePost{},
		}
		for _, p := range posts {
			output.Feed = append(output.Feed, getFeedSkeletonResponsePost{
				Post: p.URI,
			})
		}
		// Generate cursor if there are any posts, otherwise we can return an
		// empty cursor, which indicates we are at the "end" of the feed.
		if len(posts) > 0 {
			// TODO: More sophisticated cursor. Right now, if multiple posts are
			// created at the same time, a cursor based on just the created_at
			// may lead to them being omitted.
			lastPost := posts[len(posts)-1]
			output.Cursor = bluesky.FormatTime(lastPost.CreatedAt.Time)
		}

		w.WriteHeader(200)
		encoder := json.NewEncoder(w)
		_ = encoder.Encode(output)
	}
	return "/xrpc/app.bsky.feed.getFeedSkeleton", otelhttp.NewHandler(h, "get_feed_skeleton")
}
