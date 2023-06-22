package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/strideynet/bsky-furry-feed/bluesky"
	"github.com/strideynet/bsky-furry-feed/store"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"time"
)

const furryBeaconURI = "at://did:plc:kvazxresct77kgepxl7e6ay3/app.bsky.feed.post/3juew7w5lyd2n"

func postRepliesScanSource(
	ctx context.Context,
	log *zap.Logger,
	client *bluesky.Client,
	uri string,
	excludeActors []string,
) (map[string]struct{}, error) {
	prospectActors := map[string]struct{}{}
	thread, err := client.GetPostThread(ctx, uri)
	if err != nil {
		return nil, fmt.Errorf("getting post thread: %w", err)
	}
	log.Info("fetched replies to post",
		zap.String("uri", uri),
		zap.Int("reply_count", len(thread.Thread.FeedDefs_ThreadViewPost.Replies)),
	)

	for _, reply := range thread.Thread.FeedDefs_ThreadViewPost.Replies {
		actor := reply.FeedDefs_ThreadViewPost.Post.Author.Did
		// TODO: a map is going to be more performant here
		if !slices.Contains(excludeActors, actor) {
			prospectActors[actor] = struct{}{}
		}
	}

	return prospectActors, nil
}

func scanCmd(log *zap.Logger, env *environment) *cli.Command {
	return &cli.Command{
		Name:  "scan",
		Usage: "Find and add new candidate actors to add to bff",
		Action: func(cctx *cli.Context) error {
			conn, err := pgx.Connect(cctx.Context, env.dbURL)
			if err != nil {
				return fmt.Errorf("connecting to db: %w", err)
			}
			defer conn.Close(cctx.Context)

			queries := store.New(conn)

			existingActors, err := queries.ListCandidateActors(
				cctx.Context, store.NullActorStatus{},
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

				displayName := ""
				if profile.DisplayName != nil {
					displayName = *profile.DisplayName
				}

				params := store.CreateCandidateActorParams{
					DID:     profile.Did,
					Comment: fmt.Sprintf("%s (%s)", displayName, profile.Handle),
					CreatedAt: pgtype.Timestamptz{
						Time:  time.Now(),
						Valid: true,
					},
					Status: store.ActorStatusPending,
				}
				_, err = queries.CreateCandidateActor(cctx.Context, params)
				if err != nil {
					return fmt.Errorf("creating candidate actor: %w", err)
				}

				log.Info("added to queue", zap.String("did", profile.Did))
			}
			log.Info("all scanned handled")

			return nil
		},
	}
}
