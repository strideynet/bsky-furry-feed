package testing_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	. "github.com/bluesky-social/indigo/testing"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/strideynet/bsky-furry-feed/ingester"
	"github.com/strideynet/bsky-furry-feed/store"
	helper "github.com/strideynet/bsky-furry-feed/testing"
	"go.uber.org/zap"
)

var log, _ = zap.NewDevelopment()

func mustConnectDB(t *testing.T, ctx context.Context) *pgx.Conn {
	con, err := pgx.Connect(ctx, "postgres://bff:bff@localhost:5433/bff?sslmode=disable")
	assert.Nil(t, err)
	return con
}

// sets up test db and truncates all tables
func mustSetupDB(t *testing.T, ctx context.Context) {
	con := mustConnectDB(t, ctx)
	defer con.Close(ctx)

	rows, err := con.Query(ctx, "SELECT table_name FROM information_schema.tables WHERE table_schema NOT IN ('pg_catalog', 'information_schema')")
	assert.Nil(t, err)

	results, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (s string, err error) {
		err = row.Scan(&s)
		return
	})
	assert.Nil(t, err)

	tables := strings.Join(results, ", ")

	_, err = con.Exec(ctx, "TRUNCATE TABLE "+tables)
	assert.Nil(t, err)
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

	helper.SetTrialHostOnBGS(bgs, pds.RawHost())

	bob := pds.MustNewUser(t, "bob.tpds")
	furry := pds.MustNewUser(t, "furry.tpds")

	mustSetupDB(t, ctx)
	queries := store.New(mustConnectDB(t, ctx))
	cac := ingester.NewCandidateActorCache(log, queries)
	os.Setenv("BLUESKY_USERNAME", bob.DID())
	os.Setenv("BLUESKY_PASSWORD", "password")
	assert.Nil(cac.Sync(ctx))
	ca, err := cac.CreatePendingCandidateActor(ctx, furry.DID())
	assert.Nil(err)

	_, err = queries.UpdateCandidateActor(ctx, store.UpdateCandidateActorParams{
		Status: store.NullActorStatus{ActorStatus: store.ActorStatusApproved, Valid: true},
		DID:    ca.DID,
	})
	assert.Nil(err)
	assert.Nil(cac.Sync(ctx))

	ingester := ingester.NewFirehoseIngester(log, queries, cac, "ws://"+pds.RawHost()+"/xrpc/com.atproto.sync.subscribeRepos")
	go func() {
		assert.Nil(ingester.Start(ctx))
	}()

	ignoredPost := bob.Post(t, "lorem ipsum dolor sit amet")
	trackedPost := furry.Post(t, "thank u bites u")

	// ensure ingester has processed posts
	time.Sleep(time.Second)

	rows, err := mustConnectDB(t, ctx).Query(ctx, "select uri from candidate_posts")
	assert.Nil(err)
	postURIs, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (s string, err error) {
		err = row.Scan(&s)
		return
	})
	assert.Nil(err)

	assert.Equal(1, len(postURIs))
	assert.Contains(postURIs, trackedPost.Uri)
	assert.NotContains(postURIs, ignoredPost.Uri)
}
