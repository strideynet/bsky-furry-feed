package testing_test

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	. "github.com/bluesky-social/indigo/testing"
	"github.com/docker/go-connections/nat"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/strideynet/bsky-furry-feed/ingester"
	bffv1pb "github.com/strideynet/bsky-furry-feed/proto/bff/v1"
	"github.com/strideynet/bsky-furry-feed/store"
	helper "github.com/strideynet/bsky-furry-feed/testing"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

var log, _ = zap.NewDevelopment()
var postgresUrl string

func TestMain(m *testing.M) {
	flag.Parse()
	testing.Init()
	if testing.Short() {
		os.Exit(m.Run())
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	container, err := postgres.RunContainer(ctx,
		postgres.WithDatabase("bff"),
		postgres.WithUsername("bff"),
		postgres.WithPassword("bff"),
		testcontainers.WithWaitStrategy(wait.ForListeningPort(nat.Port("5432/tcp"))),
	)
	if err != nil {
		log.Fatal("could not start postgres", zap.Error(err))
	}

	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		log.Fatal("could not get postgres port", zap.Error(err))
	}
	host, err := container.Host(ctx)
	if err != nil {
		log.Fatal("could not get postgres host", zap.Error(err))
	}

	postgresUrl = fmt.Sprintf("postgres://bff:bff@%s:%d/bff?sslmode=disable", host, port.Int())

	code := m.Run()

	if err := container.Terminate(ctx); err != nil {
		log.Fatal("could not purge postgres", zap.Error(err))
	}

	os.Exit(code)
}

func mustConnectDB(t *testing.T, ctx context.Context) *pgx.Conn {
	con, err := pgx.Connect(ctx, postgresUrl)
	assert.NoError(t, err)
	return con
}

// sets up test db and truncates all tables
func mustSetupDB(t *testing.T, ctx context.Context) {
	con := mustConnectDB(t, ctx)
	defer con.Close(ctx)
	migrate, err := migrate.New("file://../store/migrations", postgresUrl)
	assert.NoError(t, err)
	err = migrate.Up()
	assert.NoError(t, err)

	rows, err := con.Query(ctx, "SELECT table_name FROM information_schema.tables WHERE table_schema NOT IN ('pg_catalog', 'information_schema')")
	assert.NoError(t, err)

	results, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (s string, err error) {
		err = row.Scan(&s)
		return
	})
	assert.NoError(t, err)

	tables := strings.Join(results, ", ")

	_, err = con.Exec(ctx, "TRUNCATE TABLE "+tables)
	assert.NoError(t, err)
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

	poolConnector := &store.DirectConnector{URI: postgresUrl}
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

	// ensure ingester has processed posts
	assert.Eventually(func() bool {
		rows, err := mustConnectDB(t, ctx).Query(ctx, "select uri from candidate_posts")
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
