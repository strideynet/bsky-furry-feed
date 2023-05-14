package feedserver

import (
	"fmt"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"net/http"
)

func didHandler(hostname string) (string, http.Handler) {
	var h http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write(
			[]byte(
				fmt.Sprintf(
					`{"@context":["https://www.w3.org/ns/did/v1"],"id":"did:web:%s","service":[{"id":"#bsky_fg","type":"BskyFeedGenerator","serviceEndpoint":"https://%s"}]}`,
					hostname,
					hostname,
				),
			),
		)
	}
	return "/.well-known/did.json", otelhttp.NewHandler(h, "get_well_known_did")
}
