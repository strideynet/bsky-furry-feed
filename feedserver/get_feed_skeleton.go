package feedserver

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	bff "github.com/strideynet/bsky-furry-feed"
	"github.com/strideynet/bsky-furry-feed/bluesky"
	"github.com/strideynet/bsky-furry-feed/store"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var feedRequestMetric = promauto.NewSummaryVec(prometheus.SummaryOpts{
	Name: "bff_feed_request_duration_seconds",
	Help: "A very rudimentary way of tracking how many feed skeletons have been requested and how long it takes to serve.",
}, []string{"feed_name", "status"})

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
	log *zap.Logger, queries *store.Queries,
) (string, http.Handler) {
	h := jsonHandler(log, func(r *http.Request) (any, error) {
		ctx := r.Context()
		start := time.Now()
		params, err := parseGetFeedSkeletonParams(r.URL)
		if err != nil {
			return nil, err
		}
		log.Debug(
			"get feed skeleton request",
			zap.String("feed", params.feed),
			zap.String("cursor", params.cursor),
			zap.Int("limit", params.limit),
		)

		// TODO: Feed "router" that directs requests to the correct
		// implementation.
		var posts []store.CandidatePost
		switch params.feed {
		case furryNewFeed, furryTestFeed:
			posts, err = getFurryNewFeed(
				ctx, queries, params.cursor, params.limit,
			)
		case furryHotFeed:
			posts, err = getFurryHotFeed(
				ctx, queries, params.cursor, params.limit,
			)
		case furryFursuitFeed:
			posts, err = getFurryNewFeedWithTag(
				ctx, queries, params.cursor, params.limit, ,
			)
		case furryArtFeed:
			posts, err = getFurryNewFeedWithTag(
				ctx, queries, params.cursor, params.limit, bff.TagArt,
			)
		case furryNSFWFeed:
			posts, err = getFurryNewFeedWithTag(
				ctx, queries, params.cursor, params.limit, bff.TagNSFW,
			)
		default:
			err = fmt.Errorf("unrecognized feed")
		}
		if err != nil {
			return nil, fmt.Errorf("fetching feed %q: %w", params.feed, err)
		}

		// Convert the selected posts to the getFeedSkeleton format
		//
		// TODO: Reintroduce pinned top-post at a later date. This should
		// only be injected if no cursor has been specified.
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

		feedRequestMetric.
			WithLabelValues(params.feed, "OK").
			Observe(time.Since(start).Seconds())
		return output, nil
	})
	return "/xrpc/app.bsky.feed.getFeedSkeleton", otelhttp.NewHandler(h, "get_feed_skeleton")
}
