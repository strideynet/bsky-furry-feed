package store

import (
	"cloud.google.com/go/cloudsqlconn"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"net"
)

var (
	// ErrNotFound indicates that no resource was found during a store call.
	ErrNotFound = fmt.Errorf("not found")
)

type DirectConnector struct {
	URI string
}

func (c *DirectConnector) poolConfig(ctx context.Context) (*pgxpool.Config, error) {
	pgxCfg, err := pgxpool.ParseConfig(c.URI)
	if err != nil {
		return nil, fmt.Errorf("parsing db url: %w", err)
	}
	return pgxCfg, nil
}

type CloudSQLConnector struct {
	Instance string
	Database string
	// TODO: Determine user from the app default service credentials
	Username string
}

func (c *CloudSQLConnector) poolConfig(ctx context.Context) (*pgxpool.Config, error) {
	d, err := cloudsqlconn.NewDialer(ctx, cloudsqlconn.WithIAMAuthN())
	if err != nil {
		return nil, fmt.Errorf("creating cloud sql dialer: %w", err)
	}
	pgxCfg, err := pgxpool.ParseConfig(fmt.Sprintf("user=%s database=%s", c.Username, c.Database))
	if err != nil {
		return nil, fmt.Errorf("parsing cloud sql config: %w", err)
	}
	pgxCfg.ConnConfig.DialFunc = func(ctx context.Context, _, _ string) (net.Conn, error) {
		return d.Dial(ctx, c.Instance)
	}
	return pgxCfg, nil
}
