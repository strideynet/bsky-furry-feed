package api

import (
	"connectrpc.com/connect"
	"context"
	"github.com/stretchr/testify/require"
	bffv1pb "github.com/strideynet/bsky-furry-feed/proto/bff/v1"
	"github.com/strideynet/bsky-furry-feed/proto/bff/v1/bffv1pbconnect"
	"github.com/strideynet/bsky-furry-feed/store"
	"net/http"
	"testing"
)

func TestAPI_UserServiceHandler_GetMe(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	harness := startAPIHarness(ctx, t)

	t.Run("existing actor", func(t *testing.T) {
		actor := harness.PDS.MustNewUser(t, "existing.tpds")
		_, err := harness.Store.CreateActor(ctx, store.CreateActorOpts{
			DID:    actor.DID(),
			Status: bffv1pb.ActorStatus_ACTOR_STATUS_APPROVED,
		})
		require.NoError(t, err)
		userSvcClient := bffv1pbconnect.NewUserServiceClient(
			http.DefaultClient,
			harness.APIAddr,
			connect.WithInterceptors(
				actorAuthInterceptor(actor),
			),
		)
		res, err := userSvcClient.GetMe(ctx, connect.NewRequest(&bffv1pb.GetMeRequest{}))
		require.NoError(t, err)
		require.Equal(t, actor.DID(), res.Msg.Actor.Did)
		require.Equal(t, bffv1pb.ActorStatus_ACTOR_STATUS_APPROVED, res.Msg.Actor.Status)
	})

	t.Run("non-existing actor", func(t *testing.T) {
		actor := harness.PDS.MustNewUser(t, "non-existing.tpds")
		userSvcClient := bffv1pbconnect.NewUserServiceClient(
			http.DefaultClient,
			harness.APIAddr,
			connect.WithInterceptors(
				actorAuthInterceptor(actor),
			),
		)
		_, err := userSvcClient.GetMe(ctx, connect.NewRequest(&bffv1pb.GetMeRequest{}))
		require.Error(t, err)
		e := err.(*connect.Error)
		require.Equal(t, connect.CodeNotFound, e.Code())
	})
}
