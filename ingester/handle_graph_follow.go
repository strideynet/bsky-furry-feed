package ingester

import (
	"context"
	"fmt"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/strideynet/bsky-furry-feed/bluesky"
	"github.com/strideynet/bsky-furry-feed/store"
	"go.uber.org/zap"
	"os"
	"time"
)

var discordWebhookGraphFollow = os.Getenv("DISCORD_WEBHOOK_GRAPH_FOLLOW")

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
