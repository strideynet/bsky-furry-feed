package ingester

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bluesky-social/indigo/api/bsky"
	bff "github.com/strideynet/bsky-furry-feed"
	"github.com/strideynet/bsky-furry-feed/bluesky"
	"github.com/strideynet/bsky-furry-feed/store"
)

func hasImage(data *bsky.FeedPost) bool {
	return data.Embed != nil && data.Embed.EmbedImages != nil && len(data.Embed.EmbedImages.Images) > 0
}

func hasKeyword(data *bsky.FeedPost, keywords ...string) bool {
	text := strings.ToLower(data.Text)
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

func isFursuitMedia(data *bsky.FeedPost) bool {
	return hasImage(data) && hasKeyword(data, "#fursuitfriday", "#fursuit")
}

func isArt(data *bsky.FeedPost) bool {
	return hasImage(data) && hasKeyword(data, "#art")
}

func isNSFW(data *bsky.FeedPost) bool {
	return hasKeyword(data, "#nsfw")
}

func isCommissionsOpen(data *bsky.FeedPost) bool {
	return hasKeyword(data, "#commsopen")
}

func (fi *FirehoseIngester) handleFeedPostCreate(
	ctx context.Context,
	repoDID string,
	recordUri string,
	data *bsky.FeedPost,
) (err error) {
	ctx, span := tracer.Start(ctx, "firehose_ingester.handle_feed_post_create")
	defer func() {
		endSpan(span, err)
	}()

	if data.Reply == nil {
		createdAt, err := bluesky.ParseTime(data.CreatedAt)
		if err != nil {
			return fmt.Errorf("parsing post time: %w", err)
		}

		// TODO: Break this out in a more extensible way
		tags := []string{}
		if isFursuitMedia(data) {
			tags = append(tags, bff.TagFursuitMedia)
		}
		if isArt(data) {
			tags = append(tags, bff.TagArt)
		}
		if isNSFW(data) {
			tags = append(tags, bff.TagNSFW)
		}
		if isCommissionsOpen(data) {
			tags = append(tags, bff.TagCommissionsOpen)
		}

		err = fi.store.CreatePost(
			ctx,
			store.CreatePostOpts{
				URI:       recordUri,
				ActorDID:  repoDID,
				CreatedAt: createdAt,
				IndexedAt: time.Now(),
				Tags:      tags,
			},
		)
		if err != nil {
			return fmt.Errorf("creating post: %w", err)
		}
	} else {
		span.AddEvent("ignoring post as it is a reply")
	}
	return nil
}

func (fi *FirehoseIngester) handleFeedPostDelete(
	ctx context.Context,
	recordUri string,
) (err error) {
	ctx, span := tracer.Start(ctx, "firehose_ingester.handle_feed_post_delete")
	defer func() {
		endSpan(span, err)
	}()

	if err := fi.store.DeletePost(
		ctx, store.DeletePostOpts{URI: recordUri},
	); err != nil {
		return fmt.Errorf("deleting post: %w", err)
	}

	return nil
}
