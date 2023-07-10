package ingester

import (
	"context"
	"fmt"
	"time"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/strideynet/bsky-furry-feed/bluesky"
	"github.com/strideynet/bsky-furry-feed/store"
)

func (fi *FirehoseIngester) handleGraphFollowCreate(
	ctx context.Context,
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
	err = fi.store.CreateFollow(
		ctx,
		store.CreateFollowOpts{
			URI:        recordUri,
			ActorDID:   repoDID,
			SubjectDID: data.Subject,
			CreatedAt:  createdAt,
			IndexedAt:  time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("creating follow: %w", err)
	}

	return nil
}

func (fi *FirehoseIngester) handleFeedFollowDelete(
	ctx context.Context,
	recordUri string,
) error {
	ctx, span := tracer.Start(ctx, "firehose_ingester.handle_feed_follow_delete")
	defer span.End()

	if err := fi.store.DeleteFollow(
		ctx, store.DeleteFollowOpts{URI: recordUri},
	); err != nil {
		return fmt.Errorf("deleting follow: %w", err)
	}

	return nil
}
