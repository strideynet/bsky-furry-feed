package testenv

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"testing"
	"unsafe"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/bluesky-social/indigo/xrpc"
	"github.com/stretchr/testify/require"
	"github.com/strideynet/bsky-furry-feed/store"

	indigoTest "github.com/bluesky-social/indigo/testing"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	ipfsLog "github.com/ipfs/go-log"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func init() {
	ipfsLog.SetAllLoggers(ipfsLog.LevelDebug)
}

// black magic to set an unexported field on the TestRelay
func setTrialHostOnRelay(trelay *indigoTest.TestRelay, rawHost string) {
	hosts := []string{rawHost}

	trialHosts := reflect.ValueOf(trelay).
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

	waitStrategy := wait.ForSQL("5432/tcp", "postgres", func(host string, port nat.Port) string {
		return fmt.Sprintf("postgres://bff:bff@%s:%d/bff?sslmode=disable", host, port.Int())
	})
	container, err := postgres.RunContainer(ctx,
		postgres.WithDatabase("bff"),
		postgres.WithUsername("bff"),
		postgres.WithPassword("bff"),
		testcontainers.WithWaitStrategy(waitStrategy),
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

	migrator, err := migrate.New("file://../store/migrations", url)
	require.NoError(t, err, "initializing migration runner")
	require.NoError(t, migrator.Up(), "applying migrations")

	return url
}

type Harness struct {
	PDS   *indigoTest.TestPDS
	Relay *indigoTest.TestRelay
	Store *store.PGXStore
}

func StartHarness(ctx context.Context, t *testing.T) *Harness {
	dbURL := StartDatabase(ctx, t)

	didr := indigoTest.TestPLC(t)

	pds := indigoTest.MustSetupPDS(t, ".tpds", didr)
	pds.Run(t)
	t.Cleanup(pds.Cleanup)

	relay := indigoTest.MustSetupRelay(t, didr, true)
	relay.Run(t)
	setTrialHostOnRelay(relay, pds.RawHost())

	pgxStore, err := store.ConnectPGXStore(
		ctx,
		slog.Default(),
		&store.DirectConnector{URI: dbURL},
	)
	require.NoError(t, err)
	t.Cleanup(pgxStore.Close)

	return &Harness{
		Relay: relay,
		PDS:   pds,
		Store: pgxStore,
	}
}
