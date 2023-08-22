package api

import (
	"fmt"
	"go.uber.org/zap"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func serverDID(hostname string) string {
	return fmt.Sprintf("did:web:%s", hostname)
}

type WebDIDService struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	ServiceEndpoint string `json:"serviceEndpoint"`
}

type WebDID struct {
	Context []string        `json:"@context"`
	ID      string          `json:"id"`
	Service []WebDIDService `json:"service"`
}

func didHandler(log *zap.Logger, hostname string) (string, http.Handler, error) {
	h := jsonHandler(log, func(r *http.Request) (any, error) {
		return WebDID{
			Context: []string{"https://www.w3.org/ns/did/v1"},
			ID:      serverDID(hostname),
			Service: []WebDIDService{
				{
					ID:              "#bsky_fg",
					Type:            "BskyFeedGenerator",
					ServiceEndpoint: fmt.Sprintf("https://%s", hostname),
				},
			},
		}, nil
	})

	return "/.well-known/did.json", otelhttp.NewHandler(h, "get_well_known_did"), nil
}
