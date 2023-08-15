package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"testing"

	"connectrpc.com/connect"
	indigoTest "github.com/bluesky-social/indigo/testing"
	"github.com/stretchr/testify/require"
	"github.com/strideynet/bsky-furry-feed/bluesky"
	"github.com/strideynet/bsky-furry-feed/feed"
	"github.com/strideynet/bsky-furry-feed/testenv"
)

func actorAuthInterceptor(actor *indigoTest.TestUser) connect.UnaryInterceptorFunc {
	token := testenv.ExtractClientFromTestUser(actor).Auth.AccessJwt
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			req.Header().Set("Authorization", "Bearer "+token)
			return next(ctx, req)
		}
	}
}

type fakeFeedService struct {
	metas []feed.Meta
}

func (m *fakeFeedService) Metas() []feed.Meta {
	return m.metas
}

func (m *fakeFeedService) GetFeedPosts(ctx context.Context, feedKey string, cursor string, limit int) (posts []feed.Post, err error) {
	return nil, fmt.Errorf("unimplemented")
}

type apiHarness struct {
	*testenv.Harness
	APIAddr string
}

func startAPIHarness(ctx context.Context, t *testing.T) *apiHarness {
	harness := testenv.StartHarness(ctx, t)

	// Create PDS user for the API taking actions as the feed
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	_ = harness.PDS.MustNewUser(t, "bff.tpds")
	srv, err := New(
		context.Background(),
		harness.Log,
		"feed.test.furryli.st",
		"",
		&fakeFeedService{
			metas: []feed.Meta{
				{
					ID:          "fake-1",
					DisplayName: "Fake",
					Description: "My Description",
				},
			},
		},
		harness.Store,
		harness.PDS.HTTPHost(),
		&bluesky.Credentials{
			Identifier: "bff.tpds",
			Password:   "password",
		},
		&AuthEngine{
			TokenValidator: BSkyTokenValidator(harness.PDS.HTTPHost()),
			ActorGetter:    harness.Store,
		},
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		srv.Close()
	})

	go func() {
		err := srv.Serve(lis)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			require.NoError(t, err)
		}
	}()

	return &apiHarness{
		Harness: harness,
		APIAddr: "http://" + lis.Addr().String(),
	}
}
