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

			queries := store.New(conn)

			out, err := bluesky.NewUnauthClient().CreateSession(
				cctx.Context, username, password,
			)
			if err != nil {
				return fmt.Errorf("authenticating: %w", err)
			}
			client := bluesky.NewClient(bluesky.AuthInfoFromCreateSession(out))

			prospectActors, err := queries.ListCandidateActors(
				cctx.Context,
				store.NullActorStatus{
					ActorStatus: store.ActorStatusPending,
					Valid:       true,
				},
			)
			if err != nil {
				return fmt.Errorf("listing candidate actors: %w", err)
			}

			for _, actor := range prospectActors {
				profile, err := client.GetProfile(cctx.Context, actor.DID)
				if err != nil {
					return fmt.Errorf("getting profile: %w, err")
				}

				displayName := ""
				if profile.DisplayName != nil {
					displayName = *profile.DisplayName
				}
				comment := fmt.Sprintf("%s (%s)", displayName, profile.Handle)

				fmt.Printf("---\n%s\n", comment)
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
					params := store.UpdateCandidateActorParams{
						DID: actor.DID,
						Status: store.NullActorStatus{
							Valid:       true,
							ActorStatus: store.ActorStatusNone,
						},
					}
					_, err := queries.UpdateCandidateActor(cctx.Context, params)
					if err != nil {
						return fmt.Errorf("creating candidate actor: %w", err)
					}

					fmt.Println("successfully rejected")
				case "add", "a":
					params := store.UpdateCandidateActorParams{
						DID: actor.DID,
						Status: store.NullActorStatus{
							Valid:       true,
							ActorStatus: store.ActorStatusApproved,
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

					_, err := queries.UpdateCandidateActor(cctx.Context, params)
					if err != nil {
						return fmt.Errorf("creating candidate actor: %w", err)
					}
					fmt.Println("successfully added")
					err = client.Follow(cctx.Context, actor.DID)
					if err != nil {
						return fmt.Errorf("following actor: %w", err)
					}
					fmt.Println("successfully followed")
				default:
					return fmt.Errorf("expected y or n but got %q", action)
				}
			}
			log.Info("all prospective actors handled")

			return nil
		},
	}
}
