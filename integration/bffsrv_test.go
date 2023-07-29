package integration_test

import (
	"context"
	"flag"
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
	"go.uber.org/zap"
)

var log, _ = zap.NewDevelopment()
var db integration.Database

func TestMain(m *testing.M) {
	flag.Parse()
	testing.Init()
	if testing.Short() {
		os.Exit(m.Run())
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	var err error
	db, err = integration.StartDatabase(ctx)

	if err != nil {
		log.Fatal("could not start integration database", zap.Error(err))
	}

	code := m.Run()

	if err := db.Close(ctx); err != nil {
		log.Fatal("could not remove test database", zap.Error(err))
	}

	os.Exit(code)
}

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	assert := assert.New(t)
	didr := TestPLC(t)

	pds := MustSetupPDS(t, ".tpds", didr)
	pds.Run(t)

	bgs := MustSetupBGS(t, didr)
	bgs.Run(t)

	integration.SetTrialHostOnBGS(bgs, pds.RawHost())

	bob := pds.MustNewUser(t, "bob.tpds")
	furry := pds.MustNewUser(t, "furry.tpds")

	assert.NoError(db.Refresh(ctx))

	poolConnector := &store.DirectConnector{URI: db.URL()}
	pgxStore, err := store.ConnectPGXStore(ctx, log.Named("store"), poolConnector)
	assert.NoError(err)
	cac := ingester.NewActorCache(log, pgxStore)
	os.Setenv("BLUESKY_USERNAME", bob.DID())
	os.Setenv("BLUESKY_PASSWORD", "password")
	_, err = pgxStore.CreateActor(ctx, store.CreateActorOpts{
		Status:  bffv1pb.ActorStatus_ACTOR_STATUS_APPROVED,
		Comment: "furry.tpds",
		DID:     furry.DID(),
	})
	assert.NoError(err)
	assert.NoError(cac.Sync(ctx))

	ingester := ingester.NewFirehoseIngester(log, pgxStore, cac, "ws://"+pds.RawHost())
	ended := false
	defer func() { ended = true }()
	go func() {
		err := ingester.Start(ctx)
		if !ended {
			assert.NoError(err)
		}
	}()

	ignoredPost := bob.Post(t, "lorem ipsum dolor sit amet")
	trackedPost := furry.Post(t, "thank u bites u")

	var postURIs []string

	con, err := db.Connect(ctx)
	assert.NoError(err)
	// ensure ingester has processed posts
	assert.Eventually(func() bool {
		rows, err := con.Query(ctx, "select uri from candidate_posts")
		assert.NoError(err)
		postURIs, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (s string, err error) {
			err = row.Scan(&s)
			return
		})
		assert.NoError(err)
		return len(postURIs) > 0
	}, time.Second, 10*time.Millisecond)

	assert.Equal(1, len(postURIs))
	assert.Contains(postURIs, trackedPost.Uri)
	assert.NotContains(postURIs, ignoredPost.Uri)
}