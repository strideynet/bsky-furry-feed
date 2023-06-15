package feedserver

import (
	"encoding/json"
	"fmt"
	"github.com/strideynet/bsky-furry-feed/store"
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
	queries *store.Queries,
	hostname string,
	listenAddr string,
) (*http.Server, error) {
	mux := &http.ServeMux{}

	didName, didHandler, err := didHandler(hostname)

	if err != nil {
		return nil, err
	}

	mux.Handle(didName, didHandler)
	mux.Handle(getFeedSkeletonHandler(log, queries))
	mux.Handle(describeFeedGeneratorHandler(log, hostname))
	mux.Handle(rootHandler(log))

	return &http.Server{
		Addr:    listenAddr,
		Handler: mux,
	}, nil
}
