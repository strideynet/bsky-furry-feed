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

func (fi *FirehoseIngester) handleFeedLikeCreate(
	ctx context.Context,
	log *zap.Logger,
	repoDID string,
	recordUri string,
	data *bsky.FeedLike,
) error {
	ctx, span := tracer.Start(ctx, "firehose_ingester.handle_feed_like_create")
	defer span.End()

	createdAt, err := bluesky.ParseTime(data.CreatedAt)
	if err != nil {
		return fmt.Errorf("parsing like time: %w", err)
	}
	err = fi.queries.CreateCandidateLike(
		ctx,
		store.CreateCandidateLikeParams{
			URI:        recordUri,
			ActorDID:   repoDID,
			SubjectURI: data.Subject.Uri,
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
		return fmt.Errorf("creating candidate like: %w", err)
	}

	return nil
}

func (fi *FirehoseIngester) handleFeedLikeDelete(
	ctx context.Context,
	log *zap.Logger,
	recordUri string,
) error {
	ctx, span := tracer.Start(ctx, "firehose_ingester.handle_feed_like_delete")
	defer span.End()

	if err := fi.queries.SoftDeleteCandidateLike(
		ctx, recordUri,
	); err != nil {
		return fmt.Errorf("deleting candidate like: %w", err)
	}

	return nil
}
