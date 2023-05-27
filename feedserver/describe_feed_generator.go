package feedserver

import (
	"fmt"
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
) (string, http.Handler) {
	h := jsonHandler(log, func(r *http.Request) (any, error) {
		res := describeFeedGeneratorResponse{
			DID: serverDID(hostname),
			Feeds: []describeFeedGeneratorResponseFeed{
				{
					URI: fmt.Sprintf(
						"at://%s/app.bsky.feed.generator/%s",
						serverDID(hostname),
						furryNewFeed,
					),
				},
				{
					URI: fmt.Sprintf(
						"at://%s/app.bsky.feed.generator/%s",
						serverDID(hostname),
						furryHotFeed,
					),
				},
				{
					URI: fmt.Sprintf(
						"at://%s/app.bsky.feed.generator/%s",
						serverDID(hostname),
						furryTestFeed,
					),
				},
			},
		}
		return res, nil
	})
	return "/xrpc/app.bsky.feed.describeFeedGenerator", otelhttp.NewHandler(h, "describe_feed_generator")
}
