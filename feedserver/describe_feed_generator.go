package feedserver

import (
	"fmt"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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
	hostname string,
) (string, http.Handler) {
	var h http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
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
		sendJSON(w, res)
	}
	return "/xrpc/app.bsky.feed.describeFeedGenerator", otelhttp.NewHandler(h, "describe_feed_generator")
}
