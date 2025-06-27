-- name: GetNewFurryNewFeed :many
SELECT
	cp.*
FROM
	candidate_posts AS cp
INNER JOIN
	candidate_actors AS ca ON cp.actor_did = ca.did
WHERE
	ca.status = 'approved'
	AND cp.is_hidden = FALSE
	AND cp.deleted_at IS NULL
	AND cp.created_at > NOW() - INTERVAL '7 day'
	AND cp.indexed_at > NOW() - INTERVAL '7 day'
	AND cp.indexed_at < sqlc.arg(cursor_timestamp)
	AND ARRAY['furry', 'furryart'] && cp.hashtags
	AND NOT ARRAY['aiart'] && cp.hashtags
	AND COALESCE(cp.has_media, cp.has_video, FALSE)
	AND NOT (ARRAY['nsfw', 'mursuit', 'murrsuit', 'nsfwfurry', 'furrynsfw'] && cp.hashtags OR ARRAY['porn', 'nudity', 'sexual'] && cp.self_labels)
ORDER BY
	cp.indexed_at DESC
LIMIT
	sqlc.arg(_limit);