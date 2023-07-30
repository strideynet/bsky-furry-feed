package integration

import (
	"context"
	"fmt"
	"github.com/bluesky-social/indigo/xrpc"
	"reflect"
	"strings"
	"unsafe"

	"github.com/bluesky-social/indigo/testing"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/ipfs/go-log"
	"github.com/jackc/pgx/v5"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func init() {
	log.SetAllLoggers(log.LevelDebug)
}

// black magic to set an unexported field on the TestBGS
func SetTrialHostOnBGS(tbgs *testing.TestBGS, rawHost string) {
	hosts := []string{rawHost}

	trialHosts := reflect.ValueOf(tbgs).
		Elem().FieldByName("tr").
		Elem().FieldByName("TrialHosts")

	reflect.NewAt(trialHosts.Type(), unsafe.Pointer(trialHosts.UnsafeAddr())).Elem().
		Set(reflect.ValueOf(hosts))
}

func ExtractClientFromTestUser(user *testing.TestUser) *xrpc.Client {
	// This isn't exposed by indigo so we have to use reflection to access the
	// client.
	val := reflect.ValueOf(user).Elem().FieldByName("client")
	iface := reflect.NewAt(val.Type(), unsafe.Pointer(val.UnsafeAddr())).Elem().Interface()
	return iface.(*xrpc.Client)
}

type Database struct {
	container *postgres.PostgresContainer
	url       string
}

func StartDatabase(ctx context.Context) (db *Database, err error) {
	container, err := postgres.RunContainer(ctx,
		postgres.WithDatabase("bff"),
		postgres.WithUsername("bff"),
		postgres.WithPassword("bff"),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp")),
	)
	if err != nil {
		return nil, fmt.Errorf("starting postgres container: %w", err)
	}

	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		return nil, fmt.Errorf("getting postgres port: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting postgres host: %w", err)
	}

	return &Database{
		container: container,
		url:       fmt.Sprintf("postgres://bff:bff@%s:%d/bff?sslmode=disable", host, port.Int()),
	}, nil
}

func (db *Database) Close(ctx context.Context) error {
	return db.container.Terminate(ctx)
}

func (db *Database) URL() string {
	return db.url
}

func (db *Database) Connect(ctx context.Context) (*pgx.Conn, error) {
	return pgx.Connect(ctx, db.URL())
}

func (db *Database) Refresh(ctx context.Context) error {
	con, err := db.Connect(ctx)
	if err != nil {
		return fmt.Errorf("connecting to test database: %w", err)
	}
	defer con.Close(ctx)

	migrator, err := migrate.New("file://../store/migrations", db.URL())
	if err != nil {
		return fmt.Errorf("initializing migration runner: %w", err)
	}
	err = migrator.Up()
	if err != nil {
		return fmt.Errorf("applying migrations: %w", err)
	}

	rows, err := con.Query(ctx, "SELECT table_name FROM information_schema.tables WHERE table_schema NOT IN ('pg_catalog', 'information_schema')")
	if err != nil {
		return fmt.Errorf("querying table names: %w", err)
	}

	results, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (s string, err error) {
		err = row.Scan(&s)
		return
	})
	if err != nil {
		return fmt.Errorf("collecting table names into array: %w", err)
	}

	tables := strings.Join(results, ", ")

	_, err = con.Exec(ctx, "TRUNCATE TABLE "+tables)
	if err != nil {
		return fmt.Errorf("truncating all tables: %v", err)
	}

	return nil
}
