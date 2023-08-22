package api

import (
	"connectrpc.com/connect"
	"context"
	"fmt"
	v1 "github.com/strideynet/bsky-furry-feed/proto/bff/v1"
)

type PublicServiceHandler struct {
	feedService feedService
}

func (p *PublicServiceHandler) ListFeeds(_ context.Context, _ *connect.Request[v1.ListFeedsRequest]) (*connect.Response[v1.ListFeedsResponse], error) {
	feeds := []*v1.Feed{}
	for _, f := range p.feedService.Metas() {
		feeds = append(feeds, &v1.Feed{
			Id: f.ID,
			// TODO(noah): Take BLUESKY_USERNAME and inject that instead of this
			// hardcoded DID in the URL.
			Link:        fmt.Sprintf("https://bsky.app/profile/did:plc:jdkvwye2lf4mingzk7qdebzc/feed/%s", f.ID),
			DisplayName: f.DisplayName,
			Description: f.Description,
			Priority:    f.Priority,
		})
	}
	return connect.NewResponse(&v1.ListFeedsResponse{
		Feeds: feeds,
	}), nil
}
