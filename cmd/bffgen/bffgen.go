package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	if err := run(); err != nil {
		os.Exit(1)
	}
}

type FeedConfig struct {
	Name     string
	BaseType string // "new" or "hot"
}

func ConditionNotHidden() string {
	return "cp.is_hidden = FALSE"
}

func ConditionHasHashtag(hashtags []string) string {
	return fmt.Sprintf(`%s && cp.hashtags`, stringSliceToArray(hashtags))
}

func ConditionNotHasHashtag(hashtags []string) string {
	return fmt.Sprintf(`NOT %s && cp.hashtags`, stringSliceToArray(hashtags))
}

func ConditionNotDeleted() string {
	return "cp.deleted_at IS NULL"
}

func ConditionActorApproved() string {
	return "ca.status = 'approved'"
}

func ConditionVideoOnly() string {
	return "COALESCE(cp.has_video, FALSE)"
}

func ConditionVideoOrImageOnly() string {
	return "COALESCE(cp.has_media, cp.has_video, FALSE)"
}

func stringSliceToArray(in []string) string {
	str := ""
	for i, h := range in {
		if i > 0 {
			str += ", "
		}
		str += fmt.Sprintf("'%s'", h)
	}
	return fmt.Sprintf("ARRAY[%s]", str)
}

func ConditionNSFWOnly() string {
	return "(ARRAY['nsfw', 'mursuit', 'murrsuit', 'nsfwfurry', 'furrynsfw'] && cp.hashtags OR ARRAY['porn', 'nudity', 'sexual'] && cp.self_labels)"
}

func ConditionSFWOnly() string {
	return "NOT " + ConditionNSFWOnly()
}

func ConditionRemoveOldCreatedAt() string {
	return "cp.created_at > NOW() - INTERVAL '7 day'"
}

func ConditionRemoveOldIndexedAt() string {
	return "cp.indexed_at > NOW() - INTERVAL '7 day'"
}

func ConditionIndexedAtCursor() string {
	return "cp.indexed_at < sqlc.arg(cursor_timestamp)"
}

func CombineConditions(conds ...string) string {
	var sb strings.Builder
	for i, c := range conds {
		sb.WriteString("\t")
		if i != 0 {
			sb.WriteString("AND ")
		}
		sb.WriteString(c)
		sb.WriteString("\n")
	}
	return sb.String()
}

func run() error {

	const baseType = "new"
	const feedName = "NewFurryNew"
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("-- name: Get%sFeed :many\n", feedName))
	if baseType == "new" {
		sb.WriteString(`SELECT
	cp.*
FROM
	candidate_posts AS cp
INNER JOIN
	candidate_actors AS ca ON cp.actor_did = ca.did
WHERE
`)
		conditions := []string{
			// Generic
			ConditionActorApproved(),
			ConditionNotHidden(),
			ConditionNotDeleted(),
			ConditionRemoveOldCreatedAt(),
			ConditionRemoveOldIndexedAt(),
			ConditionIndexedAtCursor(),
			// Specific
			ConditionHasHashtag([]string{"furry", "furryart"}),
			ConditionNotHasHashtag([]string{"aiart"}),
			ConditionVideoOrImageOnly(),
			ConditionSFWOnly(),
		}
		sb.WriteString(CombineConditions(conditions...))

		sb.WriteString(`ORDER BY
	cp.indexed_at DESC
LIMIT
	sqlc.arg(_limit);`)
	}

	f, err := os.Create(".//store/queries/" + fmt.Sprintf("feed_%s.gen.sql", feedName))
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	f.WriteString(sb.String())
	return nil
}
