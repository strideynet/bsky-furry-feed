package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func serverDID(hostname string) string {
	return fmt.Sprintf("did:web:%s", hostname)
}

func generateDIDJSON(hostname string) ([]byte, error) {
	type Object map[string]any

	did := Object{
		"@context": []string{"https://www.w3.org/ns/did/v1"},
		"id":       serverDID(hostname),
		"service": []Object{{
			"id":              "#bsky_fg",
			"type":            "BskyFeedGenerator",
			"serviceEndpoint": fmt.Sprintf("https://%s", hostname),
		}},
	}

	return json.Marshal(did)
}

func didHandler(hostname string) (string, http.Handler, error) {
	did, err := generateDIDJSON(hostname)
	if err != nil {
		return "", nil, fmt.Errorf("generating did json: %w", err)
	}

	var h http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)

		_, _ = w.Write(did)
	}

	return "/.well-known/did.json", otelhttp.NewHandler(h, "get_well_known_did"), nil
}
