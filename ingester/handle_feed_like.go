package ingester

import (
	"context"
	"fmt"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/strideynet/bsky-furry-feed/bluesky"
	"github.com/strideynet/bsky-furry-feed/store"
	"time"
)

func (fi *FirehoseIngester) handleFeedLikeCreate(
	ctx context.Context,
	repoDID string,
	recordUri string,
	data *bsky.FeedLike,
) (err error) {
	ctx, span := tracer.Start(ctx, "firehose_ingester.handle_feed_like_create")
	defer func() {
		endSpan(span, err)
	}()

	createdAt, err := bluesky.ParseTime(data.CreatedAt)
	if err != nil {
		return fmt.Errorf("parsing like time: %w", err)
	}
	err = fi.store.CreateLike(ctx, store.CreateLikeOpts{
		URI:        recordUri,
		ActorDID:   repoDID,
		SubjectURI: data.Subject.Uri,
		CreatedAt:  createdAt,
		IndexedAt:  time.Now(),
	})
	if err != nil {
		return fmt.Errorf("creating like: %w", err)
	}

	return nil
}

func (fi *FirehoseIngester) handleFeedLikeDelete(
	ctx context.Context,
	recordUri string,
) (err error) {
	ctx, span := tracer.Start(ctx, "firehose_ingester.handle_feed_like_delete")
	defer func() {
		endSpan(span, err)
	}()

	if err := fi.store.DeleteLike(
		ctx, store.DeleteLikeOpts{URI: recordUri},
	); err != nil {
		return fmt.Errorf("deleting like: %w", err)
	}

	return nil
}
