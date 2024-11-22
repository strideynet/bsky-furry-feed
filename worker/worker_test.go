package worker

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bffv1pb "github.com/strideynet/bsky-furry-feed/proto/bff/v1"
	"github.com/strideynet/bsky-furry-feed/store"
	"github.com/strideynet/bsky-furry-feed/testenv"
	typegen "github.com/whyrusleeping/cbor-gen"
)

type mockPDS struct {
	sync.Mutex
	Following map[string]bool
}

func (m *mockPDS) Follow(ctx context.Context, subjectDID string) error {
	m.Lock()
	defer m.Unlock()
	m.Following[subjectDID] = true
	return nil
}

func (m *mockPDS) Unfollow(ctx context.Context, subjectDID string) error {
	m.Lock()
	defer m.Unlock()
	delete(m.Following, subjectDID)
	return nil
}

func (m *mockPDS) isFollowed(subjectDID string) bool {
	m.Lock()
	defer m.Unlock()
	return m.Following[subjectDID]
}

type mockBGS struct {
	sync.Mutex
	records map[string]typegen.CBORMarshaler
}

func (m *mockBGS) fullPath(collection string, actorDID string, rkey string) string {
	return fmt.Sprintf("%s/%s/%s", actorDID, collection, rkey)
}

func (m *mockBGS) setRecord(collection string, actorDID string, rkey string, record typegen.CBORMarshaler) {
	m.Lock()
	defer m.Unlock()
	m.records[m.fullPath(collection, actorDID, rkey)] = record
}

func (m *mockBGS) SyncGetRecord(
	ctx context.Context, collection string, actorDID string, rkey string,
) (record typegen.CBORMarshaler, repoRev string, err error) {
	m.Lock()
	defer m.Unlock()
	record, ok := m.records[m.fullPath(collection, actorDID, rkey)]
	if !ok {
		return nil, "", &xrpc.Error{
			StatusCode: 404,
		}
	}
	return record, "1", nil
}

func strPtr(s string) *string {
	return &s
}

func TestWorker(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	dbURL := testenv.StartDatabase(ctx, t)
	pgxStore, err := store.ConnectPGXStore(
		ctx,
		slog.Default(),
		&store.DirectConnector{URI: dbURL},
	)
	require.NoError(t, err)
	t.Cleanup(pgxStore.Close)

	alreadyFollowingDID := "did:example:123"
	toFollowDID := "did:example:456"

	_, err = pgxStore.CreateActor(
		ctx,
		store.CreateActorOpts{
			DID:    alreadyFollowingDID,
			Status: bffv1pb.ActorStatus_ACTOR_STATUS_APPROVED,
		},
	)
	require.NoError(t, err)
	_, err = pgxStore.CreateActor(
		ctx,
		store.CreateActorOpts{
			DID:    toFollowDID,
			Status: bffv1pb.ActorStatus_ACTOR_STATUS_APPROVED,
		},
	)
	require.NoError(t, err)

	pds := &mockPDS{
		Following: map[string]bool{},
	}
	require.NoError(t, pds.Follow(ctx, alreadyFollowingDID))

	bgs := &mockBGS{
		records: map[string]typegen.CBORMarshaler{},
	}
	bgs.setRecord("app.bsky.actor.profile", toFollowDID, "self", &bsky.ActorProfile{
		DisplayName: strPtr("Bob Ross"),
		Description: strPtr("Happy little trees"),
	})

	w := Worker{
		log:       slog.Default(),
		store:     pgxStore,
		pdsClient: pds,
		bgsClient: bgs,
	}

	workerDone := make(chan struct{})
	go func() {
		err := w.Run(ctx)
		require.NoError(t, err)
		close(workerDone)
	}()

	t.Run("unfollow", func(t *testing.T) {
		err := pgxStore.EnqueueUnfollow(ctx, alreadyFollowingDID)
		require.NoError(t, err)

		require.EventuallyWithT(t, func(collect *assert.CollectT) {
			assert.False(
				collect,
				pds.isFollowed(alreadyFollowingDID),
				"actor should have been unfollowed")
		}, time.Second*10, time.Millisecond*100)
	})

	t.Run("follow", func(t *testing.T) {
		err := pgxStore.EnqueueFollow(ctx, toFollowDID)
		require.NoError(t, err)

		require.EventuallyWithT(t, func(collect *assert.CollectT) {
			require.True(
				collect,
				pds.isFollowed(toFollowDID),
				"actor should have been followed",
			)
			got, err := pgxStore.GetLatestActorProfile(ctx, toFollowDID)
			require.NoError(collect, err)
			require.Equal(t, "Bob Ross", got.DisplayName.String)
			require.Equal(t, "Happy little trees", got.Description.String)
		}, time.Second*10, time.Millisecond*100)
	})

	// Cancel main context and ensure our worker goes away.
	cancel()
	select {
	case <-time.After(5 * time.Second):
		assert.Fail(t, "worker failed to exit cleanly after five seconds")
	case <-workerDone:
	}
}
