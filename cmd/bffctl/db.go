package main

import (
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/strideynet/bsky-furry-feed/store"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"time"
)

const localDBURL = "postgres://bff:bff@localhost:5432/bff?sslmode=disable"

func dbCmd(log *zap.Logger) *cli.Command {
	return &cli.Command{
		Name:  "db",
		Usage: "Manage the database directly",
		Subcommands: []*cli.Command{
			{
				Name:    "candidate-repositories",
				Usage:   "Manage candidate repositories",
				Aliases: []string{"cr"},
				Subcommands: []*cli.Command{
					dbCandidateRepositoriesImportCmd(log),
				},
			},
		},
	}
}

func dbCandidateRepositoriesImportCmd(log *zap.Logger) *cli.Command {
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

			for did, candidate := range seedCandidateRepositories {
				log.Info("seeding candidate repository",
					zap.String("did", did),
					zap.Any("data", candidate),
				)
				err := db.SeedCandidateRepository(
					cctx.Context,
					store.SeedCandidateRepositoryParams{
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
		IsArtist: false,
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
}
