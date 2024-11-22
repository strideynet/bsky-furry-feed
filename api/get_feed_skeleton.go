package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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

	// Example of feed param:
	// at://did:web:feed.furryli.st/app.bsky.feed.generator/furry-chronological
	feedParam := q.Get("feed")
	if feedParam == "" {
		return nil, fmt.Errorf("no feed specified")
	}
	splitFeed := strings.Split(feedParam, "/")
	// extract the final element - we don't really care about the DID and are
	// happy to serve just based on the name.
	params.feed = splitFeed[len(splitFeed)-1]

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
	log *slog.Logger, feedService feedService,
) (string, http.Handler) {
	h := jsonHandler(log, func(r *http.Request) (any, error) {
		ctx := r.Context()

		params, err := parseGetFeedSkeletonParams(r.URL)
		if err != nil {
			return nil, err
		}
		log.Debug(
			"get feed skeleton request",
			slog.String("feed", params.feed),
			slog.String("cursor", params.cursor),
			slog.Int("limit", params.limit),
		)

		posts, err := feedService.GetFeedPosts(ctx, params.feed, params.cursor, params.limit)
		if err != nil {
			return nil, fmt.Errorf("fetching feed %q: %w", params.feed, err)
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
		// GetFeedPosts cursor if there are any posts, otherwise we can return an
		// empty cursor, which indicates we are at the "end" of the feed.
		if len(posts) > 0 {
			// TODO: More sophisticated cursor. Right now, if multiple posts are
			// created at the same time, a cursor based on just the created_at
			// may lead to them being omitted.
			lastPost := posts[len(posts)-1]
			output.Cursor = lastPost.Cursor
		}

		return output, nil
	})
	return "/xrpc/app.bsky.feed.getFeedSkeleton", otelhttp.NewHandler(h, "get_feed_skeleton")
}
