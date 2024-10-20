package ingester

import (
	"context"
	"testing"
	"time"

	"github.com/bluesky-social/indigo/api/bsky"
	indigoTest "github.com/bluesky-social/indigo/testing"
	"github.com/stretchr/testify/require"
	"github.com/strideynet/bsky-furry-feed/bluesky"
	bffv1pb "github.com/strideynet/bsky-furry-feed/proto/bff/v1"
	"github.com/strideynet/bsky-furry-feed/store"
	"github.com/strideynet/bsky-furry-feed/testenv"
)

const feedDID = "did:plc:jdkvwye2lf4mingzk7qdebzc"

func TestFirehoseIngester_FollowCreate_NeverApproved(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	harness := testenv.StartHarness(ctx, t)

	approvedFurry := harness.PDS.MustNewUser(t, "approvedFurry.tpds")
	_, err := harness.Store.CreateActor(ctx, store.CreateActorOpts{
		Status: bffv1pb.ActorStatus_ACTOR_STATUS_NONE,
		DID:    approvedFurry.DID(),
	})
	require.NoError(t, err)

	cac := NewActorCache(harness.Log, harness.Store)
	require.NoError(t, cac.Sync(ctx))
	fi := NewFirehoseIngester(
		harness.Log, harness.Store, cac, "ws://"+harness.PDS.RawHost(),
	)

	err = fi.handleGraphFollowCreate(ctx, approvedFurry.DID(), "at://"+approvedFurry.DID()+"/app.bsky.graph.follow/"+indigoTest.RandFakeCid().String(), &bsky.GraphFollow{
		LexiconTypeID: "app.bsky.graph.follow",
		CreatedAt:     bluesky.FormatTime(time.UnixMilli(0)),
		Subject:       feedDID,
	})
	require.NoError(t, err)

	actor, err := harness.Store.GetActorByDID(ctx, approvedFurry.DID())
	require.NoError(t, err)

	require.Equal(t, actor.Status, bffv1pb.ActorStatus_ACTOR_STATUS_PENDING)
}

func TestFirehoseIngester_FollowCreate_OptedOut(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	harness := testenv.StartHarness(ctx, t)

	approvedFurry := harness.PDS.MustNewUser(t, "approvedFurry.tpds")
	_, err := harness.Store.CreateActor(ctx, store.CreateActorOpts{
		Status: bffv1pb.ActorStatus_ACTOR_STATUS_OPTED_OUT,
		DID:    approvedFurry.DID(),
	})
	require.NoError(t, err)

	cac := NewActorCache(harness.Log, harness.Store)
	require.NoError(t, cac.Sync(ctx))
	fi := NewFirehoseIngester(
		harness.Log, harness.Store, cac, "ws://"+harness.PDS.RawHost(),
	)

	err = fi.handleGraphFollowCreate(ctx, approvedFurry.DID(), "at://"+approvedFurry.DID()+"/app.bsky.graph.follow/"+indigoTest.RandFakeCid().String(), &bsky.GraphFollow{
		LexiconTypeID: "app.bsky.graph.follow",
		CreatedAt:     bluesky.FormatTime(time.UnixMilli(0)),
		Subject:       feedDID,
	})
	require.NoError(t, err)

	actor, err := harness.Store.GetActorByDID(ctx, approvedFurry.DID())
	require.NoError(t, err)

	require.Equal(t, actor.Status, bffv1pb.ActorStatus_ACTOR_STATUS_APPROVED)
}

func TestFirehoseIngester_FollowCreate_Banned(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	harness := testenv.StartHarness(ctx, t)

	approvedFurry := harness.PDS.MustNewUser(t, "approvedFurry.tpds")
	_, err := harness.Store.CreateActor(ctx, store.CreateActorOpts{
		Status: bffv1pb.ActorStatus_ACTOR_STATUS_BANNED,
		DID:    approvedFurry.DID(),
	})
	require.NoError(t, err)

	cac := NewActorCache(harness.Log, harness.Store)
	require.NoError(t, cac.Sync(ctx))
	fi := NewFirehoseIngester(
		harness.Log, harness.Store, cac, "ws://"+harness.PDS.RawHost(),
	)

	err = fi.handleGraphFollowCreate(ctx, approvedFurry.DID(), "at://"+approvedFurry.DID()+"/app.bsky.graph.follow/"+indigoTest.RandFakeCid().String(), &bsky.GraphFollow{
		LexiconTypeID: "app.bsky.graph.follow",
		CreatedAt:     bluesky.FormatTime(time.UnixMilli(0)),
		Subject:       feedDID,
	})
	require.NoError(t, err)

	actor, err := harness.Store.GetActorByDID(ctx, approvedFurry.DID())
	require.NoError(t, err)

	require.Equal(t, actor.Status, bffv1pb.ActorStatus_ACTOR_STATUS_BANNED)
}

func TestFirehoseIngester_FollowDelete_OptOut(t *testing.T) {
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

	cac := NewActorCache(harness.Log, harness.Store)
	require.NoError(t, cac.Sync(ctx))
	fi := NewFirehoseIngester(
		harness.Log, harness.Store, cac, "ws://"+harness.PDS.RawHost(),
	)

	followCID := indigoTest.RandFakeCid().String()

	err = fi.handleGraphFollowCreate(ctx, approvedFurry.DID(), "at://"+approvedFurry.DID()+"/app.bsky.graph.follow/"+followCID, &bsky.GraphFollow{
		LexiconTypeID: "app.bsky.graph.follow",
		CreatedAt:     bluesky.FormatTime(time.UnixMilli(0)),
		Subject:       feedDID,
	})
	require.NoError(t, err)

	err = fi.handleGraphFollowDelete(ctx, approvedFurry.DID(), "at://"+approvedFurry.DID()+"/app.bsky.graph.follow/"+followCID)
	require.NoError(t, err)

	actor, err := harness.Store.GetActorByDID(ctx, approvedFurry.DID())
	require.NoError(t, err)

	require.Equal(t, actor.Status, bffv1pb.ActorStatus_ACTOR_STATUS_OPTED_OUT)
}

func TestFirehoseIngester_FollowDelete_Pending(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	harness := testenv.StartHarness(ctx, t)

	approvedFurry := harness.PDS.MustNewUser(t, "approvedFurry.tpds")
	_, err := harness.Store.CreateActor(ctx, store.CreateActorOpts{
		Status: bffv1pb.ActorStatus_ACTOR_STATUS_PENDING,
		DID:    approvedFurry.DID(),
	})
	require.NoError(t, err)

	cac := NewActorCache(harness.Log, harness.Store)
	require.NoError(t, cac.Sync(ctx))
	fi := NewFirehoseIngester(
		harness.Log, harness.Store, cac, "ws://"+harness.PDS.RawHost(),
	)

	followCID := indigoTest.RandFakeCid().String()

	err = fi.handleGraphFollowCreate(ctx, approvedFurry.DID(), "at://"+approvedFurry.DID()+"/app.bsky.graph.follow/"+followCID, &bsky.GraphFollow{
		LexiconTypeID: "app.bsky.graph.follow",
		CreatedAt:     bluesky.FormatTime(time.UnixMilli(0)),
		Subject:       feedDID,
	})
	require.NoError(t, err)

	err = fi.handleGraphFollowDelete(ctx, approvedFurry.DID(), "at://"+approvedFurry.DID()+"/app.bsky.graph.follow/"+followCID)
	require.NoError(t, err)

	actor, err := harness.Store.GetActorByDID(ctx, approvedFurry.DID())
	require.NoError(t, err)

	require.Equal(t, actor.Status, bffv1pb.ActorStatus_ACTOR_STATUS_NONE)
}
