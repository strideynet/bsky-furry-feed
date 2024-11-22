package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/jackc/pgx/v5"
	"github.com/strideynet/bsky-furry-feed/bfflog"
	"github.com/strideynet/bsky-furry-feed/bluesky"
	"github.com/strideynet/bsky-furry-feed/store"
	"github.com/strideynet/bsky-furry-feed/store/gen"
	typegen "github.com/whyrusleeping/cbor-gen"
	"go.uber.org/zap"
)

type pdsClient interface {
	Unfollow(ctx context.Context, subjectDID string) error
	Follow(ctx context.Context, subjectDID string) error
}

type bgsClient interface {
	SyncGetRecord(ctx context.Context, collection string, actorDID string, rkey string) (record typegen.CBORMarshaler, repoRev string, err error)
}

type Worker struct {
	log       *slog.Logger
	pdsHost   string
	pdsClient pdsClient
	bgsClient bgsClient
	store     *store.PGXStore
}

func New(
	ctx context.Context,
	log *slog.Logger,
	pdsHost string,
	bskyCredentials *bluesky.Credentials,
	pgxStore *store.PGXStore,
) (*Worker, error) {
	client, err := bluesky.ClientFromCredentials(ctx, pdsHost, bskyCredentials)
	if err != nil {
		return nil, err
	}
	bgs := &bluesky.BGSClient{}

	return &Worker{
		log:       log,
		pdsHost:   pdsHost,
		pdsClient: client,
		store:     pgxStore,
		bgsClient: bgs,
	}, nil
}

func (w *Worker) Run(ctx context.Context) error {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			task, err := w.store.GetNextFollowTask(ctx)
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			if err != nil {
				w.log.Error("loading task", bfflog.Err(err))
				continue
			}
			log := w.log.With(
				slog.Int64("task_id", task.ID),
				bfflog.ActorDID(task.ActorDID),
			)
			log.Info("processing task")

			err = w.runTask(ctx, task)
			if err != nil {
				log.Error("failed to process task", bfflog.Err(err))
				err = w.store.MarkFollowTaskAsErrored(ctx, task.ID, err)
				if err != nil {
					log.Error("failed to mark task as errored", bfflog.Err(err)))
				}

				continue
			}

			log.Info("processed task")
			err = w.store.MarkFollowTaskAsDone(ctx, task.ID)
			if err != nil {
				log.Error("marking task as done", bfflog.Err(err))
			}
		}
	}
}

func (w *Worker) runTask(ctx context.Context, task gen.FollowTask) error {
	if task.ShouldUnfollow {
		return w.pdsClient.Unfollow(ctx, task.ActorDID)
	}

	return w.updateProfileAndFollow(ctx, task.ActorDID)
}

func (w *Worker) updateProfileAndFollow(ctx context.Context, actorDid string) error {
	err := w.updateProfile(ctx, actorDid)
	if err != nil {
		return fmt.Errorf("updating profile: %w", err)
	}

	err = w.pdsClient.Follow(ctx, actorDid)
	if err != nil {
		return fmt.Errorf("following account: %w", err)
	}

	return nil
}

func (w *Worker) updateProfile(ctx context.Context, actorDid string) error {
	record, repoRev, err := w.bgsClient.SyncGetRecord(ctx, "app.bsky.actor.profile", actorDid, "self")
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

	if err := w.store.CreateLatestActorProfile(ctx, store.CreateLatestActorProfileOpts{
		ActorDID:    actorDid,
		CommitCID:   repoRev,
		CreatedAt:   time.Now(), // NOTE: The Firehose reader uses the server time but we use the local time here. This may cause staleness if the firehose gives us an older timestamp but a newer update.
		IndexedAt:   time.Now(),
		DisplayName: displayName,
		Description: description,
	}); err != nil {
		return fmt.Errorf("updating actor profile: %w", err)
	}

	return nil
}
