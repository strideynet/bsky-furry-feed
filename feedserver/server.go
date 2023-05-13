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

// https://feed-generator.skyfeed.app/xrpc/app.bsky.feed.getFeedSkeleton?feed=did:web:feed-generator.skyfeed.app/app.bsky.feed.generator/posts-with-links
func New(log *zap.Logger, st *store.Queries) *http.Server {
	mux := &http.ServeMux{}
	mux.Handle(didHandler())
	mux.Handle(getFeedSkeletonHandler(log, st))
	mux.Handle(notFoundHandler(log))

	return &http.Server{
		Addr:    ":1337",
		Handler: mux,
	}
}
