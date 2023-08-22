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

func TestModerationServiceHandler_CreateActor(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	harness := startAPIHarness(ctx, t)

	furryActor := harness.PDS.MustNewUser(t, "furry.tpds")
	modActor := harness.PDS.MustNewUser(t, "mod.tpds")
	_ = harness.PDS.MustNewUser(t, "bff.tpds")

	_, err := harness.Store.CreateActor(ctx, store.CreateActorOpts{
		DID:    modActor.DID(),
		Status: bffv1pb.ActorStatus_ACTOR_STATUS_APPROVED,
		Roles:  []string{"admin"},
	})
	require.NoError(t, err)

	modSvcClient := bffv1pbconnect.NewModerationServiceClient(
		http.DefaultClient,
		harness.APIAddr,
		connect.WithInterceptors(
			actorAuthInterceptor(modActor),
		),
	)

	_, err = modSvcClient.CreateActor(ctx, connect.NewRequest(&bffv1pb.CreateActorRequest{
		ActorDid: furryActor.DID(),
		Reason:   "im testing",
	}))
	require.NoError(t, err)

	res, err := modSvcClient.GetActor(ctx, connect.NewRequest(&bffv1pb.GetActorRequest{
		Did: furryActor.DID(),
	}))
	require.NoError(t, err)
	require.Equal(t, furryActor.DID(), res.Msg.Actor.Did)
	require.Equal(t, bffv1pb.ActorStatus_ACTOR_STATUS_NONE, res.Msg.Actor.Status)
}
