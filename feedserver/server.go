package feedserver

import (
	"github.com/strideynet/bsky-furry-feed/store"
	"go.uber.org/zap"
	"net/http"
)

type getFeedSkeletonParameters struct {
	cursor string
	limit  int
	feed   string
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
	mux.Handle(rootHandler(log))

	return &http.Server{
		Addr:    listenAddr,
		Handler: mux,
	}
}
