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

func sendJSON(w http.ResponseWriter, data any) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	encoder := json.NewEncoder(w)
	_ = encoder.Encode(data)
}

func New(
	log *zap.Logger,
	queries *store.Queries,
	hostname string,
	listenAddr string,
) *http.Server {
	mux := &http.ServeMux{}
	mux.Handle(didHandler(hostname))
	mux.Handle(getFeedSkeletonHandler(log, queries))
	mux.Handle(getCandidateActorHandler(log, queries))
	mux.Handle(describeFeedGeneratorHandler(hostname))
	mux.Handle(rootHandler(log))

	return &http.Server{
		Addr:    listenAddr,
		Handler: mux,
	}
}
