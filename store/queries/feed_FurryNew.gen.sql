-- name: GetFurryNewFeed :many
AND cp.is_hidden = FALSE
AND ARRAY['furry'] && cp.hashtags
