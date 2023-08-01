package testenv

import (
	"context"
	"fmt"
	"reflect"
	"testing"
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
	"github.com/jackc/pgx/v5"
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

func startDatabase(ctx context.Context) (close func(ctx context.Context) error, url string, err error) {
	container, err := postgres.RunContainer(ctx,
		postgres.WithDatabase("bff"),
		postgres.WithUsername("bff"),
		postgres.WithPassword("bff"),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp")),
	)
	if err != nil {
		return nil, "", fmt.Errorf("starting postgres container: %w", err)
	}

	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		return nil, "", fmt.Errorf("getting postgres port: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("getting postgres host: %w", err)
	}

	return func(ctx context.Context) error {
		return container.Terminate(ctx)
	}, fmt.Sprintf("postgres://bff:bff@%s:%d/bff?sslmode=disable", host, port.Int()), nil
}

func runMigrations(dbURL string) error {
	migrator, err := migrate.New("file://../store/migrations", dbURL)
	if err != nil {
		return fmt.Errorf("initializing migration runner: %w", err)
	}
	err = migrator.Up()
	if err != nil {
		return fmt.Errorf("applying migrations: %w", err)
	}

	return nil
}

type Harness struct {
	DBConn *pgx.Conn
	PDS    *indigoTest.TestPDS
	BGS    *indigoTest.TestBGS
	Log    *zap.Logger
	Store  *store.PGXStore
}

func StartHarness(ctx context.Context, t *testing.T) *Harness {
	log := zaptest.NewLogger(t)

	dbClose, dbURL, err := startDatabase(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, dbClose(context.Background()))
	})

	conn, err := pgx.Connect(ctx, dbURL)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, conn.Close(context.Background()))
	})
	require.NoError(t, runMigrations(dbURL))

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
		DBConn: conn,
		BGS:    bgs,
		PDS:    pds,
		Log:    log,
		Store:  pgxStore,
	}
}
