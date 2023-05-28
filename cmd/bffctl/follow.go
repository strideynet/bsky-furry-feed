package main

import (
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/strideynet/bsky-furry-feed/bluesky"
	"github.com/strideynet/bsky-furry-feed/store"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

func followCmd(log *zap.Logger, env *environment) *cli.Command {
	return &cli.Command{
		Name:  "follow",
		Usage: "Follows candidate actors with the logged-in account",
		Action: func(cctx *cli.Context) error {
			conn, err := pgx.Connect(cctx.Context, env.dbURL)
			if err != nil {
				return fmt.Errorf("connecting to db: %w", err)
			}
			defer conn.Close(cctx.Context)

			queries := store.New(conn)

			subjects, err := queries.ListCandidateActors(cctx.Context, store.NullActorStatus{
				ActorStatus: store.ActorStatusApproved,
				Valid:       true,
			})
			if err != nil {
				return fmt.Errorf("listing candidate actors: %w", err)
			}

			out, err := bluesky.NewUnauthClient().CreateSession(
				cctx.Context, username, password,
			)
			if err != nil {
				return fmt.Errorf("authenticating: %w", err)
			}

			client := bluesky.NewClient(bluesky.AuthInfoFromCreateSession(out))

			for _, subject := range subjects {
				err := client.Follow(cctx.Context, subject.DID)
				if err != nil {
					return fmt.Errorf("following actor: %w", err)
				}
			}

			log.Info("all prospective actors handled")

			return nil
		},
	}
}
