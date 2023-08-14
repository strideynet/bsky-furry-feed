package api

import (
	"connectrpc.com/connect"
	"context"
	"github.com/stretchr/testify/require"
	"github.com/strideynet/bsky-furry-feed/feed"
	v1 "github.com/strideynet/bsky-furry-feed/proto/bff/v1"
	"testing"
)

type fakeMetaSourcer struct {
	metas []feed.Meta
}

func (m *fakeMetaSourcer) Metas() []feed.Meta {
	return m.metas
}

func TestPublicServiceHandler_ListFeeds(t *testing.T) {
	h := PublicServiceHandler{feedMetaSourcer: &fakeMetaSourcer{
		metas: []feed.Meta{
			{
				ID:          "foo",
				DisplayName: "Foo Feed",
				Description: "Descriptiones",
				Priority:    9000,
			},
		},
	}}

	res, err := h.ListFeeds(context.Background(), connect.NewRequest(&v1.ListFeedsRequest{}))
	require.NoError(t, err)
	require.Equal(t, &v1.ListFeedsResponse{
		Feeds: []*v1.Feed{
			{
				Id:          "foo",
				DisplayName: "Foo Feed",
				Description: "Descriptiones",
				Priority:    9000,
				Link:        "https://bsky.app/profile/did:plc:jdkvwye2lf4mingzk7qdebzc/feed/foo",
			},
		},
	}, res.Msg)
}
