package main

import (
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/strideynet/bsky-furry-feed/bluesky"
	"github.com/strideynet/bsky-furry-feed/store"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"strings"
	"time"
)

func dbCmd(log *zap.Logger, env *environment) *cli.Command {
	return &cli.Command{
		Name:  "db",
		Usage: "Manage the database directly",
		Subcommands: []*cli.Command{
			{
				Name:    "candidate-repositories",
				Usage:   "Manage candidate repositories",
				Aliases: []string{"cr"},
				Subcommands: []*cli.Command{
					dbCandidateRepositoriesList(log, env),
					dbCandidateRepositoriesSeedCmd(log, env),
					dbCandidateRepositoriesAddCmd(log, env),
				},
			},
		},
	}
}

func dbCandidateRepositoriesList(log *zap.Logger, env *environment) *cli.Command {
	return &cli.Command{
		Name:  "ls",
		Usage: "Listcandidate repositories",
		Action: func(cctx *cli.Context) error {
			conn, err := pgx.Connect(cctx.Context, env.dbURL)
			if err != nil {
				return err
			}
			defer conn.Close(cctx.Context)

			db := store.New(conn)
			repos, err := db.ListCandidateRepositories(cctx.Context)
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

func dbCandidateRepositoriesSeedCmd(log *zap.Logger, env *environment) *cli.Command {
	return &cli.Command{
		Name:  "seed",
		Usage: "Seed the default set of candidate repositories",
		Action: func(cctx *cli.Context) error {
			conn, err := pgx.Connect(cctx.Context, env.dbURL)
			if err != nil {
				return err
			}
			defer conn.Close(cctx.Context)

			db := store.New(conn)

			log.Info("seed candidates", zap.Int("count", len(seedCandidateRepositories)))
			for did, candidate := range seedCandidateRepositories {
				log.Info("seeding candidate repository",
					zap.String("did", did),
					zap.Any("data", candidate),
				)
				err := db.CreateCandidateRepository(
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

func dbCandidateRepositoriesAddCmd(log *zap.Logger, env *environment) *cli.Command {
	handle := ""
	name := ""
	isArtist := false
	return &cli.Command{
		Name:  "add",
		Usage: "Adds a new candidate repository",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "handle",
				Required:    true,
				Destination: &handle,
			},
			&cli.StringFlag{
				Name:        "name",
				Required:    true,
				Destination: &name,
			},
			&cli.BoolFlag{
				Name:        "artist",
				Destination: &isArtist,
			},
		},
		Action: func(cctx *cli.Context) error {
			conn, err := pgx.Connect(cctx.Context, env.dbURL)
			if err != nil {
				return err
			}
			defer conn.Close(cctx.Context)

			client := bluesky.NewClient()
			did, err := client.ResolveHandle(cctx.Context, handle)
			if err != nil {
				return err
			}
			log.Info("found did", zap.String("did", did.Did))

			db := store.New(conn)

			params := store.CreateCandidateRepositoryParams{
				DID: did.Did,
				CreatedAt: pgtype.Timestamptz{
					Time:  time.Now(),
					Valid: true,
				},
				IsArtist: isArtist,
				Comment:  fmt.Sprintf("%s (%s)", name, handle),
			}
			log.Info("adding candidate repository",
				zap.Any("data", params),
			)
			err = db.CreateCandidateRepository(
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
			log.Info("added candidate repository")

			return nil
		},
	}
}

var seedCandidateRepositories = map[string]struct {
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
