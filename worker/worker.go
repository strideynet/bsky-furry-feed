package worker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/jackc/pgx/v5"
	"github.com/strideynet/bsky-furry-feed/bluesky"
	"github.com/strideynet/bsky-furry-feed/store"
	"go.uber.org/zap"
)

func Start(
	ctx context.Context,
	log *zap.Logger,
	pdsHost string,
	bskyCredentials *bluesky.Credentials,
	pgxStore *store.PGXStore,
) error {
	client, err := bluesky.ClientFromCredentials(ctx, pdsHost, bskyCredentials)
	if err != nil {
		return err
	}
	bgsClient := bluesky.BGSClient{}

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			task, err := pgxStore.GetNextFollowTask(ctx)
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}

			log.Info("processing task", zap.Int64("id", task.ID))

			if task.ShouldUnfollow {
				err = client.Unfollow(ctx, task.ActorDID)
			} else {
				err = client.Follow(ctx, task.ActorDID)

				if err != nil {
					record, repoRev, err := bgsClient.SyncGetRecord(ctx, "app.bsky.actor.profile", task.ActorDID, "self")
					if err != nil {
						if err2 := (&xrpc.Error{}); !errors.As(err, &err2) || err2.StatusCode != 404 {
							return fmt.Errorf("getting profile: %w", err)
						}
						record = nil
					}

					var profile *bsky.ActorProfile
					if record != nil {
						switch record := record.(type) {
						case *bsky.ActorProfile:
							profile = record
						default:
							return fmt.Errorf("expected *bsky.ActorProfile, got %T", record)
						}
					}

					displayName := ""
					description := ""

					if profile != nil {
						if profile.DisplayName != nil {
							displayName = *profile.DisplayName
						}

						if profile.Description != nil {
							description = *profile.Description
						}
					}

					if err := pgxStore.CreateLatestActorProfile(ctx, store.CreateLatestActorProfileOpts{
						ActorDID:    task.ActorDID,
						CommitCID:   repoRev,
						CreatedAt:   time.Now(), // NOTE: The Firehose reader uses the server time but we use the local time here. This may cause staleness if the firehose gives us an older timestamp but a newer update.
						IndexedAt:   time.Now(),
						DisplayName: displayName,
						Description: description,
					}); err != nil {
						return fmt.Errorf("updating actor profile: %w", err)
					}

				}
			}

			if err != nil {
				log.Error("failed to process task", zap.Int64("id", task.ID), zap.Error(err))
				err = pgxStore.MarkFollowTaskAsErrored(ctx, task.ID, err)
				if err != nil {
					return fmt.Errorf("marking task %d as errored: %w", task.ID, err)
				}

				continue
			}

			log.Info("processed task", zap.Int64("id", task.ID))
			err = pgxStore.MarkFollowTaskAsDone(ctx, task.ID)
			if err != nil {
				return fmt.Errorf("marking follow task as done: %w", err)
			}
		}
	}
}
