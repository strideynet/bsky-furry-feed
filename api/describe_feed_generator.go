package api

import (
	"fmt"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type describeFeedGeneratorResponseFeed struct {
	URI string `json:"uri"`
}

type describeFeedGeneratorResponse struct {
	DID   string                              `json:"did"`
	Feeds []describeFeedGeneratorResponseFeed `json:"feeds"`
}

func describeFeedGeneratorHandler(
	log *slog.Logger,
	hostname string,
	feedService feedService,
) (string, http.Handler) {
	feedURI := func(feedName string) string {
		return fmt.Sprintf(
			// TODO(noah): This should be returning the profile that owns the
			// content not the serverDID
			"at://%s/app.bsky.feed.generator/%s",
			serverDID(hostname),
			feedName,
		)
	}

	feeds := []describeFeedGeneratorResponseFeed{}
	for _, meta := range feedService.Metas() {
		feeds = append(feeds, describeFeedGeneratorResponseFeed{
			URI: feedURI(meta.ID),
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
