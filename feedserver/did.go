package feedserver

import (
	"fmt"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"net/http"
)

func serverDID(hostname string) string {
	return fmt.Sprintf("did:web:%s", hostname)
}

func didHandler(hostname string) (string, http.Handler) {
	var h http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		// TODO: Make a struct for this rather than Sprintfing a json string.
		_, _ = w.Write(
			[]byte(
				fmt.Sprintf(
					`{"@context":["https://www.w3.org/ns/did/v1"],"id":"%s","service":[{"id":"#bsky_fg","type":"BskyFeedGenerator","serviceEndpoint":"https://%s"}]}`,
					serverDID(hostname),
					hostname,
				),
			),
		)
	}
	return "/.well-known/did.json", otelhttp.NewHandler(h, "get_well_known_did")
}
