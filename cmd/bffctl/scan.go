package main

import (
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/strideynet/bsky-furry-feed/bluesky"
	"github.com/strideynet/bsky-furry-feed/store"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"os"
)

type stateFile struct {
	previouslyRejected map[string]string
}

// TODO: Have a `login` and `logout` command that persists auth state to disk.
var username = os.Getenv("BSKY_USERNAME")
var password = os.Getenv("BSKY_PASSWORD")

func scanCmd(log *zap.Logger, env *environment) *cli.Command {
	return &cli.Command{
		Name:  "scan",
		Usage: "Find and add new candidate repositories to add to bff",
		Action: func(cctx *cli.Context) error {
			conn, err := pgx.Connect(cctx.Context, env.dbURL)
			if err != nil {
				return err
			}
			defer conn.Close(cctx.Context)

			_ = store.New(conn)

			out, err := bluesky.NewUnauthClient().CreateSession(cctx.Context, username, password)
			if err != nil {
				return fmt.Errorf("failed to authenticate: %w", err)
			}

			client := bluesky.NewClient(bluesky.AuthInfoFromCreateSession(out))

			return nil
		},
	}
}
