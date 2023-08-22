package api

import (
	"connectrpc.com/connect"
	"context"
	"errors"
	indigoTest "github.com/bluesky-social/indigo/testing"
	"github.com/stretchr/testify/require"
	"github.com/strideynet/bsky-furry-feed/bluesky"
	"github.com/strideynet/bsky-furry-feed/feed"
	"github.com/strideynet/bsky-furry-feed/testenv"
	"net"
	"net/http"
	"testing"
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

type apiHarness struct {
	*testenv.Harness
	APIAddr string
}

func startAPIHarness(ctx context.Context, t *testing.T) *apiHarness {
	harness := testenv.StartHarness(ctx, t)

	// Create PDS user for the API taking actions as the feed
	_ = harness.PDS.MustNewUser(t, "bff.tpds")
	srv, err := New(
		harness.Log,
		"",
		"",
		&feed.Service{},
		harness.Store,
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
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

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
