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
					dbCandidateActorsSeedCmd(log, env),
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

func dbCandidateActorsSeedCmd(log *zap.Logger, env *environment) *cli.Command {
	return &cli.Command{
		Name:  "seed",
		Usage: "Seed the default set of candidate actors",
		Action: func(cctx *cli.Context) error {
			conn, err := pgx.Connect(cctx.Context, env.dbURL)
			if err != nil {
				return err
			}
			defer conn.Close(cctx.Context)

			db := gen.New(conn)

			log.Info("seed candidates", zap.Int("count", len(seedCandidateActors)))
			for did, candidate := range seedCandidateActors {
				log.Info("seeding candidate actor",
					zap.String("did", did),
					zap.Any("data", candidate),
				)
				_, err := db.CreateCandidateActor(
					cctx.Context,
					gen.CreateCandidateActorParams{
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
					if strings.Contains(err.Error(), "duplicate key") {
						log.Warn(
							"already exists, no action taken",
							zap.String("did", did),
						)
					} else {
						return err
					}
				}
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

			client, err := getBlueskyClient(cctx.Context)
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

var seedCandidateActors = map[string]struct {
	Comment  string
	IsArtist bool
}{
	"did:plc:dllwm3fafh66ktjofzxhylwk": {
		Comment:  "Noah (ottr.sh)",
		IsArtist: false,
	},
	"did:plc:jt43524ltn23seg5v3qhurwt": {
		Comment:  "vilk (vilk.pub)",
		IsArtist: false,
	},
	"did:plc:ouytv644apqbu2pm7fnp7qrj": {
		Comment:  "Newton (newton.dog)",
		IsArtist: true,
	},
	"did:plc:hjzrjs7sewv6nmratpoeavtp": {
		Comment:  "Kepler",
		IsArtist: false,
	},
	"did:plc:ojw5gcvjs44m7dl5zrzeb4i3": {
		Comment:  "Rend (dingo.bsky.social)",
		IsArtist: false,
	},
	"did:plc:ggg7g6gcc65lzwqvqqpa2mik": {
		Comment:  "Concoction (concoction.bsky.social)",
		IsArtist: true,
	},
	"did:plc:sfvpv6dfrug3rnjewn7gyx62": {
		Comment:  "qdot (buttplug.engineer)",
		IsArtist: false,
	},
	"did:plc:o74zbazekchwk2v4twee4ekb": {
		Comment:  "kio (kio.dev)",
		IsArtist: true,
	},
	"did:plc:rgbf6ph3eki5lffvrs6syf4w": {
		Comment:  "cael (cael.tech)",
		IsArtist: false,
	},
	"did:plc:wtfep3izymr6ot4tywoqcydc": {
		Comment:  "adam (snowfox.gay)",
		IsArtist: false,
	},
	"did:plc:6aikzgasri74fypm4h3qfvui": {
		Comment:  "havokhusky (havok.bark.supply)",
		IsArtist: false,
	},
	"did:plc:rjawzv3m7smnyaiq62mrqpok": {
		Comment:  "frank (lickmypa.ws)",
		IsArtist: false,
	},
	"did:plc:f3ynrkwdfe7m5ffvxd5pxf4f": {
		Comment:  "lobo (lupine.agency)",
		IsArtist: false,
	},
	"did:plc:q6j66z2z7hkwjssiq7zzz3ej": {
		Comment:  "zenith (pawgge.rs)",
		IsArtist: false,
	},
	"did:plc:inze6wrmsm7pjl7yta3oig77": {
		Comment:  "videah (videah.net)",
		IsArtist: false,
	},
	"did:plc:74vecggtrogqfv3fdmflhhtq": {
		Comment:  "wuff (imjusta.dog)",
		IsArtist: false,
	},
	"did:plc:cuo7esdirjyrw53uffaczjmt": {
		Comment:  "aero (aero.bsky.social)",
		IsArtist: false,
	},
	"did:plc:wherpiavw4rekzkmc6egfy4y": {
		Comment:  "lio (pogcha.mp)",
		IsArtist: false,
	},
	"did:plc:w5a2nvvmdatnyb2cyijwwk3v": {
		Comment:  "reese (reese.bsky.social)",
		IsArtist: false,
	},
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

			client, err := getBlueskyClient(cctx.Context)
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
