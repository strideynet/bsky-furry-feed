package ingester

import (
	"bytes"
	"context"
	"fmt"
	"github.com/bluesky-social/indigo/api/bsky"
	"go.uber.org/zap"
	"net/http"
	"os"
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

	if fi.crc.GetByDID(data.Subject) != nil {
		// We aren't interested in repositories we already track.
		return nil
	}
	if discordWebhookGraphFollow == "" {
		log.Warn("no webhook configured for graph follow")
		return nil
	}

	var jsonStr = []byte(fmt.Sprintf(`
{
    "username": "bff-system",
    "content": "**Furry follow?**\n**Source repository:** %s \n**Followed:** https://psky.app/profile/%s"
}`, repoDID, data.Subject))
	req, err := http.NewRequest("POST", discordWebhookGraphFollow, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("sending discord webhook: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
