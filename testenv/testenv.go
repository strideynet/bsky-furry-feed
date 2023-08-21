package testenv

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/bluesky-social/indigo/xrpc"
	"github.com/stretchr/testify/require"
	"github.com/strideynet/bsky-furry-feed/store"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	indigoTest "github.com/bluesky-social/indigo/testing"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	ipfsLog "github.com/ipfs/go-log"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func init() {
	ipfsLog.SetAllLoggers(ipfsLog.LevelDebug)
}

// black magic to set an unexported field on the TestBGS
func setTrialHostOnBGS(tbgs *indigoTest.TestBGS, rawHost string) {
	hosts := []string{rawHost}

	trialHosts := reflect.ValueOf(tbgs).
		Elem().FieldByName("tr").
		Elem().FieldByName("TrialHosts")

	reflect.NewAt(
		trialHosts.Type(),
		unsafe.Pointer(trialHosts.UnsafeAddr())).Elem().Set(reflect.ValueOf(hosts))
}

func ExtractClientFromTestUser(user *indigoTest.TestUser) *xrpc.Client {
	// This isn't exposed by indigo so we have to use reflection to access the
	// client.
	val := reflect.ValueOf(user).Elem().FieldByName("client")
	iface := reflect.NewAt(val.Type(), unsafe.Pointer(val.UnsafeAddr())).Elem().Interface()
	return iface.(*xrpc.Client)
}

func StartDatabase(ctx context.Context, t *testing.T) (url string) {
	t.Helper()

	container, err := postgres.RunContainer(ctx,
		postgres.WithDatabase("bff"),
		postgres.WithUsername("bff"),
		postgres.WithPassword("bff"),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp")),
	)
	require.NoError(t, err, "starting postgres container")
	t.Cleanup(func() {
		require.NoError(t, container.Terminate(context.Background()))
	})

	port, err := container.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err, "getting postgres port")

	host, err := container.Host(ctx)
	require.NoError(t, err, "getting postgres host")

	url = fmt.Sprintf("postgres://bff:bff@%s:%d/bff?sslmode=disable", host, port.Int())

	// Use EventuallyWithT as it can take a few additional seconds for the db
	// to become stably healthy.
	var migrator *migrate.Migrate
	require.EventuallyWithT(t, func(t *assert.CollectT) {
		migrator, err = migrate.New("file://../store/migrations", url)
		assert.NoError(t, err, "initializing migration runner")
	}, time.Second*5, time.Millisecond*250)

	require.NoError(t, migrator.Up(), "applying migrations")

	return url
}

type Harness struct {
	PDS   *indigoTest.TestPDS
	BGS   *indigoTest.TestBGS
	Log   *zap.Logger
	Store *store.PGXStore
}

func StartHarness(ctx context.Context, t *testing.T) *Harness {
	log := zaptest.NewLogger(t)

	dbURL := StartDatabase(ctx, t)

	didr := indigoTest.TestPLC(t)

	pds := indigoTest.MustSetupPDS(t, ".tpds", didr)
	pds.Run(t)
	t.Cleanup(pds.Cleanup)

	bgs := indigoTest.MustSetupBGS(t, didr)
	bgs.Run(t)
	setTrialHostOnBGS(bgs, pds.RawHost())

	pgxStore, err := store.ConnectPGXStore(
		ctx,
		log.Named("store"),
		&store.DirectConnector{URI: dbURL},
	)
	require.NoError(t, err)
	t.Cleanup(pgxStore.Close)

	return &Harness{
		BGS:   bgs,
		PDS:   pds,
		Log:   log,
		Store: pgxStore,
	}
}
