package feedserver

import (
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

func New(
	log *zap.Logger,
	queries *store.Queries,
	hostname string,
	listenAddr string,
) *http.Server {
	mux := &http.ServeMux{}
	mux.Handle(didHandler(hostname))
	mux.Handle(getFeedSkeletonHandler(log, queries))
	mux.Handle(getCandidateRepositoryHandler(log, queries))
	mux.Handle(rootHandler(log))

	return &http.Server{
		Addr:    listenAddr,
		Handler: mux,
	}
}
