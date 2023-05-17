package ingester

import (
	"context"
	"fmt"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/strideynet/bsky-furry-feed/bluesky"
	"github.com/strideynet/bsky-furry-feed/store"
	"go.uber.org/zap"
	"time"
)

func (fi *FirehoseIngester) handleFeedPostCreate(
	ctx context.Context,
	log *zap.Logger,
	repoDID string,
	recordUri string,
	data *bsky.FeedPost,
) error {
	ctx, span := tracer.Start(ctx, "firehose_ingester.handle_feed_post_create")
	defer span.End()
	if data.Reply == nil {
		createdAt, err := bluesky.ParseTime(data.CreatedAt)
		if err != nil {
			return fmt.Errorf("parsing post time: %w", err)
		}
		err = fi.queries.CreateCandidatePost(
			ctx,
			store.CreateCandidatePostParams{
				URI:           recordUri,
				RepositoryDID: repoDID,
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
			return fmt.Errorf("creating candidate post: %w", err)
		}
	} else {
		log.Info("ignoring reply")
	}
	return nil
}
