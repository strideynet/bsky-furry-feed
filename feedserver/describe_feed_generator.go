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
	feedURI := func(feedName string) string {
		return fmt.Sprintf(
			"at://%s/app.bsky.feed.generator/%s",
			serverDID(hostname),
			feedName,
		)
	}

	h := jsonHandler(log, func(r *http.Request) (any, error) {
		res := describeFeedGeneratorResponse{
			DID: serverDID(hostname),
			Feeds: []describeFeedGeneratorResponseFeed{
				// TODO: Iterate over some central feed registry
				{
					URI: feedURI(furryNewFeed),
				},
				{
					URI: feedURI(furryHotFeed),
				},
				{
					URI: feedURI(furryTestFeed),
				},
				{
					URI: feedURI(furryFursuitFeed),
				},
				{
					URI: feedURI(furryArtFeed),
				},
				{
					URI: feedURI(furryNSFWFeed),
				},
			},
		}
		return res, nil
	})
	return "/xrpc/app.bsky.feed.describeFeedGenerator", otelhttp.NewHandler(h, "describe_feed_generator")
}
