package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/strideynet/bsky-furry-feed/store/gen"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

func dbCmd(log *zap.Logger, env *environment) *cli.Command {
	return &cli.Command{
		Name:  "db",
		Usage: "Manage the database directly",
		Subcommands: []*cli.Command{
			{
				Name:    "candidate-actors",
				Usage:   "Manage candidate actors",
				Aliases: []string{"ca"},
				Subcommands: []*cli.Command{
					dbCandidateActorsList(log, env),
					dbCandidateActorsAddCmd(log, env),
					dbCandidateActorsBackfillProfiles(log, env),
				},
			},
		},
	}
}

func dbCandidateActorsList(log *zap.Logger, env *environment) *cli.Command {
	return &cli.Command{
		Name:  "ls",
		Usage: "List candidate actors",
		Action: func(cctx *cli.Context) error {
			conn, err := pgx.Connect(cctx.Context, env.dbURL)
			if err != nil {
				return err
			}
			defer conn.Close(cctx.Context)

			db := gen.New(conn)
			repos, err := db.ListCandidateActors(cctx.Context, gen.NullActorStatus{})
			if err != nil {
				return err
			}
			for _, r := range repos {
				log.Info("repo", zap.Any("data", r))
			}
			return nil
		},
	}
}

func dbCandidateActorsAddCmd(log *zap.Logger, env *environment) *cli.Command {
	handle := ""
	isArtist := false
	shouldFollow := false
	return &cli.Command{
		Name:  "add",
		Usage: "Adds a new candidate actor",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "handle",
				Required:    true,
				Destination: &handle,
			},
			&cli.BoolFlag{
				Name:        "artist",
				Destination: &isArtist,
			},
			&cli.BoolFlag{
				Name:        "follow",
				Destination: &shouldFollow,
				Usage:       "follows the actor after adding them",
			},
		},
		Action: func(cctx *cli.Context) error {
			conn, err := pgx.Connect(cctx.Context, env.dbURL)
			if err != nil {
				return err
			}
			defer conn.Close(cctx.Context)

			client, err := getBlueskyClient(cctx.Context, env)
			if err != nil {
				return err
			}

			did, err := client.ResolveHandle(cctx.Context, handle)
			if err != nil {
				return fmt.Errorf("resolving handle: %w", err)
			}
			log.Info("found did", zap.String("did", did.Did))

			db := gen.New(conn)

			params := gen.CreateCandidateActorParams{
				DID: did.Did,
				CreatedAt: pgtype.Timestamptz{
					Time:  time.Now(),
					Valid: true,
				},
				Roles:    []string{},
				IsArtist: isArtist,
				Comment:  handle,
				Status:   gen.ActorStatusApproved,
			}
			log.Info("adding candidate actor",
				zap.Any("data", params),
			)
			_, err = db.CreateCandidateActor(
				cctx.Context,
				params,
			)
			if err != nil {
				if strings.Contains(err.Error(), "duplicate key") {
					log.Warn(
						"already exists, no action taken",
						zap.String("did", did.Did),
					)
				} else {
					return err
				}
			}
			log.Info("successfully added")
			if shouldFollow {
				if err := client.Follow(cctx.Context, did.Did); err != nil {
					return fmt.Errorf("following actor: %w", err)
				}
				log.Info("successfully followed")
			}
			return nil
		},
	}
}

func dbCandidateActorsBackfillProfiles(log *zap.Logger, env *environment) *cli.Command {
	return &cli.Command{
		Name:  "backfill-profiles",
		Usage: "Backfill profiles for all actors missing profiles",
		Action: func(cctx *cli.Context) error {
			conn, err := pgx.Connect(cctx.Context, env.dbURL)
			if err != nil {
				return err
			}
			defer conn.Close(cctx.Context)

			client, err := getBlueskyClient(cctx.Context, env)
			if err != nil {
				return err
			}

			db := gen.New(conn)
			repos, err := db.ListCandidateActorsRequiringProfileBackfill(cctx.Context)
			if err != nil {
				return err
			}
			for _, r := range repos {
				// TODO: Does this need throttling?
				head, err := client.GetHead(cctx.Context, r.DID)
				if err != nil {
					return fmt.Errorf("getting head: %w", err)
				}

				record, err := client.GetRecord(cctx.Context, "app.bsky.actor.profile", head, r.DID, "self")
				if err != nil {
					return fmt.Errorf("getting profile: %w", err)
				}

				var profile *bsky.ActorProfile
				switch record := record.(type) {
				case *bsky.ActorProfile:
					profile = record
				default:
					return fmt.Errorf("expected *bsky.ActorProfile, got %T", record)
				}

				displayName := ""
				if profile.DisplayName != nil {
					displayName = *profile.DisplayName
				}

				description := ""
				if profile.Description != nil {
					description = *profile.Description
				}

				params := gen.CreateLatestActorProfileParams{
					ActorDID:  r.DID,
					CommitCID: head.String(),
					CreatedAt: pgtype.Timestamptz{
						Valid: true,
						// NOTE: The Firehose reader uses the server time but we use the local time here. This may cause staleness if the firehose gives us an older timestamp but a newer update.
						Time: time.Now(),
					},
					IndexedAt: pgtype.Timestamptz{
						Valid: true,
						Time:  time.Now(),
					},
					DisplayName: pgtype.Text{
						Valid:  true,
						String: displayName,
					},
					Description: pgtype.Text{
						Valid:  true,
						String: description,
					},
				}
				log.Info("backfilling candidate actor profile",
					zap.Any("data", params),
				)
				if err := db.CreateLatestActorProfile(cctx.Context, params); err != nil {
					return err
				}

			}
			return nil
		},
	}
}
