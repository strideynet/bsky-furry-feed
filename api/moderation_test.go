package api_test

import (
	"context"
	"errors"
	"github.com/bufbuild/connect-go"
	"github.com/stretchr/testify/require"
	"github.com/strideynet/bsky-furry-feed/api"
	"github.com/strideynet/bsky-furry-feed/bluesky"
	"github.com/strideynet/bsky-furry-feed/feed"
	bffv1pb "github.com/strideynet/bsky-furry-feed/proto/bff/v1"
	"github.com/strideynet/bsky-furry-feed/proto/bff/v1/bffv1pbconnect"
	"github.com/strideynet/bsky-furry-feed/testenv"
	"net"
	"net/http"
	"testing"
)

func TestAPI_CreateActor(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	harness := testenv.StartHarness(ctx, t)

	furryActor := harness.PDS.MustNewUser(t, "furry.tpds")
	modActor := harness.PDS.MustNewUser(t, "mod.tpds")
	_ = harness.PDS.MustNewUser(t, "bff.tpds")

	srv, err := api.New(
		harness.Log,
		"",
		"",
		&feed.Service{},
		harness.Store,
		&bluesky.Credentials{
			Identifier: "bff.tpds",
			Password:   "password",
		},
		&api.AuthEngine{
			PDSHost: harness.PDS.HTTPHost(),
			ModeratorDIDs: []string{
				modActor.DID(),
			},
		},
	)
	require.NoError(t, err)
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer lis.Close()
	go func() {
		err := srv.Serve(lis)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			require.NoError(t, err)
		}
	}()
	defer srv.Close()

	modActorToken := testenv.ExtractClientFromTestUser(modActor).Auth.AccessJwt
	modSvc := bffv1pbconnect.NewModerationServiceClient(
		http.DefaultClient,
		"http://"+lis.Addr().String(),
		connect.WithInterceptors(connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
			return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
				req.Header().Set("Authorization", "Bearer "+modActorToken)
				return next(ctx, req)
			}
		})),
	)

	_, err = modSvc.CreateActor(ctx, connect.NewRequest(&bffv1pb.CreateActorRequest{
		ActorDid: furryActor.DID(),
		Reason:   "im testing",
	}))
	require.NoError(t, err)

	res, err := modSvc.GetActor(ctx, connect.NewRequest(&bffv1pb.GetActorRequest{
		Did: furryActor.DID(),
	}))
	require.NoError(t, err)
	require.Equal(t, furryActor.DID(), res.Msg.Actor.Did)
	require.Equal(t, bffv1pb.ActorStatus_ACTOR_STATUS_NONE, res.Msg.Actor.Status)
}
