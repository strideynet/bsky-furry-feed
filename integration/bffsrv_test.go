package integration_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/strideynet/bsky-furry-feed/api"
	"github.com/strideynet/bsky-furry-feed/bluesky"
	"go.uber.org/zap/zaptest"
	"os"
	"testing"
	"time"

	. "github.com/bluesky-social/indigo/testing"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/strideynet/bsky-furry-feed/ingester"
	"github.com/strideynet/bsky-furry-feed/integration"
	bffv1pb "github.com/strideynet/bsky-furry-feed/proto/bff/v1"
	"github.com/strideynet/bsky-furry-feed/store"
)

func TestIngester(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	t.Parallel()

	log := zaptest.NewLogger(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := integration.StartDatabase(ctx)
	defer db.Close(context.Background())
	require.NoError(t, err)
	didr := TestPLC(t)
	pds := MustSetupPDS(t, ".tpds", didr)
	pds.Run(t)
	defer pds.Cleanup()
	bgs := MustSetupBGS(t, didr)
	bgs.Run(t)
	integration.SetTrialHostOnBGS(bgs, pds.RawHost())

	bob := pds.MustNewUser(t, "bob.tpds")
	furry := pds.MustNewUser(t, "furry.tpds")

	require.NoError(t, db.Refresh(ctx))

	poolConnector := &store.DirectConnector{URI: db.URL()}
	pgxStore, err := store.ConnectPGXStore(ctx, log.Named("store"), poolConnector)
	require.NoError(t, err)
	cac := ingester.NewActorCache(log, pgxStore)
	require.NoError(t, os.Setenv("BLUESKY_USERNAME", bob.DID()))
	require.NoError(t, os.Setenv("BLUESKY_PASSWORD", "password"))
	_, err = pgxStore.CreateActor(ctx, store.CreateActorOpts{
		Status:  bffv1pb.ActorStatus_ACTOR_STATUS_APPROVED,
		Comment: "furry.tpds",
		DID:     furry.DID(),
	})
	require.NoError(t, err)
	require.NoError(t, cac.Sync(ctx))

	fi := ingester.NewFirehoseIngester(log, pgxStore, cac, "ws://"+pds.RawHost())
	ended := false
	defer func() { ended = true }()
	go func() {
		err := fi.Start(ctx)
		if !ended {
			require.NoError(t, err)
		}
	}()

	ignoredPost := bob.Post(t, "lorem ipsum dolor sit amet")
	trackedPost := furry.Post(t, "thank u bites u")

	var postURIs []string

	con, err := db.Connect(ctx)
	require.NoError(t, err)
	// ensure ingester has processed posts
	require.Eventually(t, func() bool {
		rows, err := con.Query(ctx, "select uri from candidate_posts")
		require.NoError(t, err)
		postURIs, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (s string, err error) {
			err = row.Scan(&s)
			return
		})
		require.NoError(t, err)
		return len(postURIs) > 0
	}, time.Second, 10*time.Millisecond)

	assert.Equal(t, 1, len(postURIs))
	assert.Contains(t, postURIs, trackedPost.Uri)
	assert.NotContains(t, postURIs, ignoredPost.Uri)
}

func TestAPI_CreateActor(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	t.Parallel()

	log := zaptest.NewLogger(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// TODO: Extract all of this to a testing harness
	db, err := integration.StartDatabase(ctx)
	defer db.Close(context.Background())
	require.NoError(t, err)
	didr := TestPLC(t)
	pds := MustSetupPDS(t, ".tpds", didr)
	pds.Run(t)
	defer pds.Cleanup()
	bgs := MustSetupBGS(t, didr)
	bgs.Run(t)
	integration.SetTrialHostOnBGS(bgs, pds.RawHost())

	modActor := pds.MustNewUser(t, "mod.tpds")
	_ = pds.MustNewUser(t, "bff.tpds")

	poolConnector := &store.DirectConnector{URI: db.URL()}
	pgxStore, err := store.ConnectPGXStore(ctx, log.Named("store"), poolConnector)
	require.NoError(t, err)
	srv, err := api.New(
		log,
		"",
		"",
		nil,
		pgxStore,
		&bluesky.Credentials{
			Identifier: "bff.tpds",
			Password:   "password",
		},
		&api.AuthEngine{
			PDSHost: pds.HTTPHost(),
			ModeratorDIDs: []string{
				modActor.DID(),
			},
		},
	)
	require.NoError(t, err)
	require.NoError(t, srv.ListenAndServe())
	defer srv.Close()

}
