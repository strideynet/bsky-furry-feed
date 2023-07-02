package api

import (
	"encoding/json"
	"fmt"
	"github.com/strideynet/bsky-furry-feed/feed"
	"go.uber.org/zap"
	"net/http"
)

func handleErr(w http.ResponseWriter, log *zap.Logger, err error) {
	log.Error("failed to handle request", zap.Error(err))
	w.WriteHeader(500)
	w.Write([]byte(fmt.Sprintf("failed to handle request: %s", err)))
}

func jsonHandler(log *zap.Logger, h func(r *http.Request) (any, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		respBody, err := h(r)
		if err != nil {
			handleErr(w, log, err)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(200)
		encoder := json.NewEncoder(w)
		_ = encoder.Encode(respBody)
	}
}

func New(
	log *zap.Logger,
	hostname string,
	listenAddr string,
	feedRegistry *feed.Service,
) (*http.Server, error) {
	mux := &http.ServeMux{}

	didEndpointPath, didHandler, err := didHandler(hostname)
	if err != nil {
		return nil, fmt.Errorf("creating did handler: %w", err)
	}
	mux.Handle(didEndpointPath, didHandler)
	mux.Handle(getFeedSkeletonHandler(log, feedRegistry))
	mux.Handle(describeFeedGeneratorHandler(log, hostname, feedRegistry))
	mux.Handle(rootHandler(log))

	return &http.Server{
		Addr:    listenAddr,
		Handler: mux,
	}, nil
}
