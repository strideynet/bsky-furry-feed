package ingester

import (
	"context"
	"fmt"
	"time"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/ipfs/go-cid"
	"github.com/strideynet/bsky-furry-feed/store"
)

func (fi *FirehoseIngester) handleActorProfileUpdate(
	ctx context.Context,
	repoDID string,
	commitCID cid.Cid,
	recordUri string,
	createdAt time.Time,
	data *bsky.ActorProfile,
) (err error) {
	ctx, span := tracer.Start(ctx, "firehose_ingester.handle_actor_profile_update")
	defer func() {
		endSpan(span, err)
	}()

	displayName := ""
	if data.DisplayName != nil {
		displayName = *data.DisplayName
	}

	description := ""
	if data.Description != nil {
		description = *data.Description
	}

	selfLabels := []string{}
	if data.Labels != nil && data.Labels.LabelDefs_SelfLabels != nil {
		for _, label := range data.Labels.LabelDefs_SelfLabels.Values {
			selfLabels = append(selfLabels, label.Val)
		}
	}

	err = fi.store.CreateLatestActorProfile(
		ctx,
		store.CreateLatestActorProfileOpts{
			ActorDID:    repoDID,
			CommitCID:   commitCID.String(),
			CreatedAt:   createdAt,
			IndexedAt:   time.Now(),
			DisplayName: displayName,
			Description: description,
			SelfLabels:  selfLabels,
		},
	)
	if err != nil {
		return fmt.Errorf("updating profile: %w", err)
	}

	return nil
}

func (fi *FirehoseIngester) handleActorProfileDelete(
	ctx context.Context,
	recordUri string,
) (err error) {
	ctx, span := tracer.Start(ctx, "firehose_ingester.handle_actor_profile_delete")
	_ = ctx
	defer func() {
		endSpan(span, err)
	}()

	// TODO: Maybe purge the actor from the database?

	return nil
}
