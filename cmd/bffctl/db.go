package main

import (
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	bff "github.com/strideynet/bsky-furry-feed"
	"github.com/strideynet/bsky-furry-feed/store"
	"github.com/urfave/cli/v2"
	"time"
)

const localDBURL = "postgres://bff:bff@localhost:5432/bff?sslmode=disable"

func dbCmd() *cli.Command {
	return &cli.Command{
		Name:  "db",
		Usage: "Manage the database directly",
		Subcommands: []*cli.Command{
			{
				Name:    "candidate_repositories",
				Usage:   "Manage candidate repositories",
				Aliases: []string{"cr"},
				Subcommands: []*cli.Command{
					dbCandidateRepositoriesImportCmd(),
				},
			},
		},
	}
}

func dbCandidateRepositoriesImportCmd() *cli.Command {
	return &cli.Command{
		Name:  "import",
		Usage: "Import the default set of candidate repositories",
		Action: func(cctx *cli.Context) error {
			conn, err := pgx.Connect(cctx.Context, localDBURL)
			if err != nil {
				return err
			}
			defer conn.Close(cctx.Context)

			db := store.New(conn)

			for did, candidate := range bff.NewStaticCandidateUsers() {
				_, err := db.CreateCandidateRepository(
					cctx.Context,
					store.CreateCandidateRepositoryParams{
						DID: did,
						CreatedAt: pgtype.Timestamptz{
							Time:  time.Now(),
							Valid: true,
						},
						IsArtist: candidate.IsArtist,
						Comment:  candidate.Comment,
					},
				)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
}
