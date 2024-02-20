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
) (err error) {
	ctx, span := tracer.Start(ctx, "firehose_ingester.handle_graph_follow_create")
	defer func() {
		endSpan(span, err)
	}()

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

	if fi.IsFurryFeedDID(data.Subject) {
		if err := fi.actorCache.OptInOrMarkPending(ctx, repoDID); err != nil {
			return fmt.Errorf("opting in: %w", err)
		}
	}

	return nil
}

func (fi *FirehoseIngester) handleGraphFollowDelete(
	ctx context.Context,
	repoDID string,
	recordUri string,
) (err error) {
	ctx, span := tracer.Start(ctx, "firehose_ingester.handle_feed_follow_delete")
	defer func() {
		endSpan(span, err)
	}()

	subjectDID, err := fi.store.DeleteFollow(
		ctx, store.DeleteFollowOpts{URI: recordUri},
	)

	if err != nil {
		return fmt.Errorf("deleting follow: %w", err)
	}

	if fi.IsFurryFeedDID(subjectDID) {
		if err := fi.actorCache.OptOutOrForget(ctx, repoDID); err != nil {
			return fmt.Errorf("opting out: %w", err)
		}
	}

	return nil
}
