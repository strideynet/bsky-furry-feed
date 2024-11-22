package ingester

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/bluesky-social/indigo/repo"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bffv1pb "github.com/strideynet/bsky-furry-feed/proto/bff/v1"
	"github.com/strideynet/bsky-furry-feed/store"
	"github.com/strideynet/bsky-furry-feed/testenv"
)

func TestFirehoseIngester_ActorProfiles(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	harness := testenv.StartHarness(ctx, t)

	approvedFurry := harness.PDS.MustNewUser(t, "approvedFurry.tpds")
	_, err := harness.Store.CreateActor(ctx, store.CreateActorOpts{
		Status: bffv1pb.ActorStatus_ACTOR_STATUS_APPROVED,
		DID:    approvedFurry.DID(),
	})
	require.NoError(t, err)

	cac := NewActorCache(slog.Default(), harness.Store)
	require.NoError(t, cac.Sync(ctx))
	fi := NewFirehoseIngester(
		slog.Default(), harness.Store, cac, "ws://"+harness.PDS.RawHost(),
	)

	{
		displayName := "some furry"
		description := "hewwo :3"
		err = fi.handleActorProfileUpdate(ctx, approvedFurry.DID(), repo.NextTID(), "at://"+approvedFurry.DID()+"/app.bsky.actor.profile/self", time.UnixMilli(0), &bsky.ActorProfile{
			LexiconTypeID: "app.bsky.actor.profile",
			DisplayName:   &displayName,
			Description:   &description,
		})
		require.NoError(t, err)

		ap, err := harness.Store.GetLatestActorProfile(ctx, approvedFurry.DID())
		require.NoError(t, err)

		assert.Equal(t, ap.DisplayName.String, "some furry")
		assert.Equal(t, ap.Description.String, "hewwo :3")
	}

	{
		displayName := "some other furry"
		description := "hewwo >:3"

		err = fi.handleActorProfileUpdate(ctx, approvedFurry.DID(), repo.NextTID(), "at://"+approvedFurry.DID()+"/app.bsky.actor.profile/self", time.UnixMilli(1), &bsky.ActorProfile{
			LexiconTypeID: "app.bsky.actor.profile",
			DisplayName:   &displayName,
			Description:   &description,
		})
		require.NoError(t, err)

		ap, err := harness.Store.GetLatestActorProfile(ctx, approvedFurry.DID())
		require.NoError(t, err)

		assert.Equal(t, ap.DisplayName.String, "some other furry")
		assert.Equal(t, ap.Description.String, "hewwo >:3")
	}

	aps, err := harness.Store.GetActorProfileHistory(ctx, approvedFurry.DID())
	require.NoError(t, err)

	assert.Equal(t, aps[0].DisplayName.String, "some other furry")
	assert.Equal(t, aps[0].Description.String, "hewwo >:3")

	assert.Equal(t, aps[1].DisplayName.String, "some furry")
	assert.Equal(t, aps[1].Description.String, "hewwo :3")
}
