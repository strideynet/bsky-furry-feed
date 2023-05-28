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

			existingActors, err := queries.ListCandidateActors(
				cctx.Context,
				store.NullActorStatus{
					ActorStatus: store.ActorStatusPending,
					Valid:       true,
				},
			)
			if err != nil {
				return fmt.Errorf("listing candidate actors: %w", err)
			}
			// Exclude actors that we already know about.
			// TODO: a map is going to be more performant here
			excludeActorDIDs := []string{}
			for _, actor := range existingActors {
				excludeActorDIDs = append(excludeActorDIDs, actor.DID)
			}

			out, err := bluesky.NewUnauthClient().CreateSession(
				cctx.Context, username, password,
			)
			if err != nil {
				return fmt.Errorf("authenticating: %w", err)
			}

			client := bluesky.NewClient(bluesky.AuthInfoFromCreateSession(out))
			// TODO: Inject existing repositories/reject history as exclude actors
			prospectActors, err := postRepliesScanSource(
				cctx.Context, log, client, furryBeaconURI, excludeActorDIDs,
			)
			if err != nil {
				return fmt.Errorf("fetching prospect actors: %w", err)
			}
			log.Info("scan phase complete", zap.Int("found_count", len(prospectActors)))

			for actor := range prospectActors {
				profile, err := client.GetProfile(cctx.Context, actor)
				if err != nil {
					return fmt.Errorf("getting profile: %w, err")
				}

				displayName := "none"
				if profile.DisplayName != nil {
					displayName = *profile.DisplayName
				}
				desc := "none"
				if profile.Description != nil {
					desc = *profile.Description
				}

				fmt.Printf("---\n%s (%s)\n", displayName, profile.Handle)
				fmt.Printf("link: https://bsky.app/profile/%s\n", profile.Did)
				fmt.Printf("desc:\n%s\n", desc)
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
				case "add", "a":
					params := store.CreateCandidateActorParams{
						DID:     profile.Did,
						Comment: fmt.Sprintf("%s (%s)", displayName, profile.Handle),
						CreatedAt: pgtype.Timestamptz{
							Time:  time.Now(),
							Valid: true,
						},
						Status: store.ActorStatusApproved,
					}
					fmt.Printf("is this account an artist [y/n]: ")

					isArtist := ""
					_, err = fmt.Scanln(&isArtist)
					if err != nil {
						return fmt.Errorf("scanning user input: %w", err)
					}
					switch strings.ToLower(isArtist) {
					case "y":
						params.IsArtist = true
					case "n":
						params.IsArtist = false
					default:
						return fmt.Errorf("expected y or n but got %q", isArtist)
					}

					err := queries.CreateCandidateActor(cctx.Context, params)
					if err != nil {
						return fmt.Errorf("creating candidate actor: %w", err)
					}

					fmt.Println("successfully added")
				default:
					return fmt.Errorf("expected y or n but got %q", action)
				}
			}
			log.Info("all prospective actors handled")

			return nil
		},
	}
}
