package ingester

import (
	"context"
	"fmt"
	"time"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/strideynet/bsky-furry-feed/bluesky"
	"github.com/strideynet/bsky-furry-feed/store"
	"go.uber.org/zap"
)

func (fi *FirehoseIngester) handleGraphFollowCreate(
	ctx context.Context,
	log *zap.Logger,
	repoDID string,
	recordUri string,
	data *bsky.GraphFollow,
) error {
	ctx, span := tracer.Start(ctx, "firehose_ingester.handle_graph_follow_create")
	defer span.End()

	createdAt, err := bluesky.ParseTime(data.CreatedAt)
	if err != nil {
		return fmt.Errorf("parsing follow time: %w", err)
	}
	err = fi.queries.CreateCandidateFollow(
		ctx,
		store.CreateCandidateFollowParams{
			URI:        recordUri,
			ActorDID:   repoDID,
			SubjectDid: data.Subject,
			CreatedAt: pgtype.Timestamptz{
				Time:  createdAt,
				Valid: true,
			},
			IndexedAt: pgtype.Timestamptz{
				Time:  time.Now(),
				Valid: true,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("creating candidate follow: %w", err)
	}

	return nil
}

func (fi *FirehoseIngester) handleFeedFollowDelete(
	ctx context.Context,
	log *zap.Logger,
	recordUri string,
) error {
	ctx, span := tracer.Start(ctx, "firehose_ingester.handle_feed_follow_delete")
	defer span.End()

	if err := fi.queries.SoftDeleteCandidateFollow(
		ctx, recordUri,
	); err != nil {
		return fmt.Errorf("deleting candidate follow: %w", err)
	}

	return nil
}
