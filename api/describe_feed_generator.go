package api

import (
	"fmt"
	"github.com/strideynet/bsky-furry-feed/feed"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
	"net/http"
)

type describeFeedGeneratorResponseFeed struct {
	URI string `json:"uri"`
}

type describeFeedGeneratorResponse struct {
	DID   string                              `json:"did"`
	Feeds []describeFeedGeneratorResponseFeed `json:"feeds"`
}

func describeFeedGeneratorHandler(
	log *zap.Logger,
	hostname string,
	registry *feed.Service,
) (string, http.Handler) {
	feedURI := func(feedName string) string {
		return fmt.Sprintf(
			"at://%s/app.bsky.feed.generator/%s",
			serverDID(hostname),
			feedName,
		)
	}

	feeds := []describeFeedGeneratorResponseFeed{}
	for _, id := range registry.IDs() {
		feeds = append(feeds, describeFeedGeneratorResponseFeed{
			URI: feedURI(id),
		})
	}

	h := jsonHandler(log, func(r *http.Request) (any, error) {
		res := describeFeedGeneratorResponse{
			DID:   serverDID(hostname),
			Feeds: feeds,
		}
		return res, nil
	})
	return "/xrpc/app.bsky.feed.describeFeedGenerator", otelhttp.NewHandler(h, "describe_feed_generator")
}
