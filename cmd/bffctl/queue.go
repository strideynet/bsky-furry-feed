package main

import (
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/strideynet/bsky-furry-feed/store/gen"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"strings"
)

func queueCmd(log *zap.Logger, env *environment) *cli.Command {
	return &cli.Command{
		Name:  "queue",
		Usage: "Process entries in the queue",
		Action: func(cctx *cli.Context) error {
			conn, err := pgx.Connect(cctx.Context, env.dbURL)
			if err != nil {
				return fmt.Errorf("connecting to db: %w", err)
			}
			defer conn.Close(cctx.Context)

			queries := gen.New()

			client, err := getBlueskyClient(cctx.Context)
			if err != nil {
				return err
			}

			prospectActors, err := queries.ListCandidateActors(
				cctx.Context,
				conn,
				gen.NullActorStatus{
					ActorStatus: gen.ActorStatusPending,
					Valid:       true,
				},
			)
			if err != nil {
				return fmt.Errorf("listing candidate actors: %w", err)
			}

			for i, actor := range prospectActors {
				profile, err := client.GetProfile(cctx.Context, actor.DID)
				if err != nil {
					return fmt.Errorf("getting profile: %w", err)
				}

				displayName := ""
				if profile.DisplayName != nil {
					displayName = *profile.DisplayName
				}
				comment := fmt.Sprintf("%s (%s)", displayName, profile.Handle)

				fmt.Printf(
					"---\n[%d/%d] %s\n",
					i+1,
					len(prospectActors),
					comment,
				)
				if profile.Description != nil {
					fmt.Println()
					fmt.Println(*profile.Description)
					fmt.Println()
				}
				fmt.Printf("link: https://bsky.app/profile/%s\n", actor.DID)
				fmt.Printf("(a)dd, (r)eject, (s)kip, (q)uit: ")
				action := ""
				_, err = fmt.Scanln(&action)
				if err != nil {
					return fmt.Errorf("scanning user input: %w", err)
				}

				switch strings.ToLower(action) {
				case "skip", "s":
					continue
				case "quit", "q":
					return nil
				case "reject", "r":
					fmt.Println("rejecting...")
					params := gen.UpdateCandidateActorParams{
						DID: actor.DID,
						Status: gen.NullActorStatus{
							Valid:       true,
							ActorStatus: gen.ActorStatusNone,
						},
					}
					_, err := queries.UpdateCandidateActor(cctx.Context, conn, params)
					if err != nil {
						return fmt.Errorf("creating candidate actor: %w", err)
					}

					fmt.Println("successfully rejected")
				case "add", "a":
					params := gen.UpdateCandidateActorParams{
						DID: actor.DID,
						Status: gen.NullActorStatus{
							Valid:       true,
							ActorStatus: gen.ActorStatusApproved,
						},
						IsArtist: pgtype.Bool{
							Valid: true,
							// Actual value will be filled below.
						},
						Comment: pgtype.Text{
							Valid:  true,
							String: comment,
						},
					}
					fmt.Printf("is this account an artist [y/n]: ")

					isArtist := ""
					_, err = fmt.Scanln(&isArtist)
					if err != nil {
						return fmt.Errorf("scanning user input: %w", err)
					}
					switch strings.ToLower(isArtist) {
					case "y":
						params.IsArtist.Bool = true
					case "n":
						params.IsArtist.Bool = false
					default:
						return fmt.Errorf("expected y or n but got %q", isArtist)
					}

					log.Info("adding")
					_, err := queries.UpdateCandidateActor(cctx.Context, conn, params)
					if err != nil {
						return fmt.Errorf("creating candidate actor: %w", err)
					}
					log.Info("successfully added")
					log.Info("following")
					err = client.Follow(cctx.Context, actor.DID)
					if err != nil {
						return fmt.Errorf("following actor: %w", err)
					}
					log.Info("successfully followed")
				default:
					return fmt.Errorf("expected y or n but got %q", action)
				}
			}
			log.Info("all prospective actors handled")

			return nil
		},
	}
}
