package ingester

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bffv1pb "github.com/strideynet/bsky-furry-feed/proto/bff/v1"
	"github.com/strideynet/bsky-furry-feed/store"
	"github.com/strideynet/bsky-furry-feed/testenv"
)

func TestCandidateActorCache(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	dbURI := testenv.StartDatabase(ctx, t)
	pgxStore, err := store.ConnectPGXStore(
		ctx,
		slog.Default(),
		&store.DirectConnector{URI: dbURI},
	)
	require.NoError(t, err)
	t.Cleanup(pgxStore.Close)

	preApprovedDID := "pre-approved"
	_, err = pgxStore.CreateActor(ctx, store.CreateActorOpts{
		Status: bffv1pb.ActorStatus_ACTOR_STATUS_APPROVED,
		DID:    "pre-approved",
	})
	require.NoError(t, err)

	clock := clockwork.NewFakeClock()
	cac := NewActorCache(slog.Default(), pgxStore)
	cac.clock = clock
	require.NoError(t, cac.Sync(ctx))

	// Check for the actor we created before the sync
	got := cac.GetByDID(preApprovedDID)
	require.NotNil(t, got)
	require.Equal(t, preApprovedDID, got.Did)
	require.Equal(t, bffv1pb.ActorStatus_ACTOR_STATUS_APPROVED, got.Status)

	// Check pending actor creation
	createdDID := "created-pending"
	require.NoError(t, cac.CreatePendingCandidateActor(ctx, createdDID))
	got = cac.GetByDID(createdDID)
	require.NotNil(t, got)
	require.Equal(t, createdDID, got.Did)
	require.Equal(t, bffv1pb.ActorStatus_ACTOR_STATUS_PENDING, got.Status)

	// Actually start the CAC
	cacContext, cacStop := context.WithCancel(ctx)
	cacDone := make(chan struct{})
	go func() {
		require.NoError(t, cac.Start(cacContext))
		close(cacDone)
	}()

	// Approve the previously created pending actor
	_, err = pgxStore.UpdateActor(ctx, store.UpdateActorOpts{
		DID:          createdDID,
		UpdateStatus: bffv1pb.ActorStatus_ACTOR_STATUS_APPROVED,
	})
	require.NoError(t, err)

	// Ensure this doesn't take immediate effect
	got = cac.GetByDID(createdDID)
	require.NotNil(t, got)
	require.Equal(t, createdDID, got.Did)
	require.Equal(t, bffv1pb.ActorStatus_ACTOR_STATUS_PENDING, got.Status)

	// Magically advance time
	clock.Advance(time.Minute * 2)

	// Ensure that within a few milliseconds this changes to approved
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		got = cac.GetByDID(createdDID)
		if !assert.NotNil(t, got) {
			return
		}
		assert.Equal(t, createdDID, got.Did)
		assert.Equal(t, bffv1pb.ActorStatus_ACTOR_STATUS_APPROVED, got.Status)
	}, time.Second, time.Millisecond*100)

	// Check shut down is clean (e.g it doesn't hang)
	cacStop()
	select {
	case <-time.After(time.Second * 5):
		require.FailNow(t, "cac did not exit within deadline")
	case <-cacDone:
	}
}
