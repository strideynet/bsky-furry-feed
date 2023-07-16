package integration

import (
	"context"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/bluesky-social/indigo/testing"
	"github.com/docker/go-connections/nat"
	"github.com/ipfs/go-log"
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

type Database struct {
	container *postgres.PostgresContainer
	url       string
}

func StartDatabase(ctx context.Context) (db Database, err error) {
	container, err := postgres.RunContainer(ctx,
		postgres.WithDatabase("bff"),
		postgres.WithUsername("bff"),
		postgres.WithPassword("bff"),
		testcontainers.WithWaitStrategy(wait.ForListeningPort(nat.Port("5432/tcp"))),
	)

	if err != nil {
		err = fmt.Errorf("could not start postgres: %w", err)
		return
	}

	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		err = fmt.Errorf("could not get postgres port: %w", err)
		return
	}

	host, err := container.Host(ctx)
	if err != nil {
		err = fmt.Errorf("could not get postgres host: %w", err)
		return
	}

	db.container = container
	db.url = fmt.Sprintf("postgres://bff:bff@%s:%d/bff?sslmode=disable", host, port.Int())
	return
}

func (db Database) Close(ctx context.Context) error {
	return db.container.Terminate(ctx)
}

func (db Database) URL() string {
	return db.url
}
