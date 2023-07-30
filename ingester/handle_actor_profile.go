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
	recordUri string,
	recordCid cid.Cid,
	updatedAt time.Time,
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

	err = fi.store.UpdateActorProfile(
		ctx,
		store.UpdateActorProfileOpts{
			DID:         repoDID,
			CID:         recordCid,
			UpdatedAt:   updatedAt,
			DisplayName: displayName,
			Description: description,
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
	defer func() {
		endSpan(span, err)
	}()

	// TODO: Maybe purge the actor from the database?

	return nil
}
