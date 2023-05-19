package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/strideynet/bsky-furry-feed/bluesky"
	"github.com/strideynet/bsky-furry-feed/store"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"os"
)

type stateFile struct {
	previouslyRejected map[string]string
}

// TODO: Have a `login` and `logout` command that persists auth state to disk.
var username = os.Getenv("BSKY_USERNAME")
var password = os.Getenv("BSKY_PASSWORD")

const furryBeaconURI = "at://did:plc:kvazxresct77kgepxl7e6ay3/app.bsky.feed.post/3juew7w5lyd2n"

func postRepliesScanSource(
	ctx context.Context,
	log *zap.Logger,
	client *bluesky.Client,
	uri string,
	excludeActors []string,
) ([]string, error) {
	prospectActors := []string{}
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
			prospectActors = append(prospectActors, actor)
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

			store := store.New(conn)

			existingActors, err := store.ListCandidateActors(cctx.Context)
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

			for _, actor := range prospectActors {
				profile, err := client.GetProfile(cctx.Context, actor)
				if err != nil {
					return fmt.Errorf("getting profile: %w, err")
				}
				fmt.Printf("%s (%s) - https://bsky.app/profile/%s\n", *profile.DisplayName, profile.Handle, profile.Did)
			}

			return nil
		},
	}
}
