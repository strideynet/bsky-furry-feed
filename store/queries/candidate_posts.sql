-- name: CreateCandidatePost :exec
INSERT INTO
    candidate_posts (uri, actor_did, created_at, indexed_at, hashtags, has_media, raw)
VALUES
    ($1, $2, $3, $4, $5, $6, $7);

-- name: SoftDeleteCandidatePost :exec
UPDATE
    candidate_posts
SET
    deleted_at = NOW()
WHERE
    uri = $1;

-- name: GetFurryNewFeed :many
SELECT
    cp.*
FROM
    candidate_posts cp
        INNER JOIN candidate_actors ca ON cp.actor_did = ca.did
WHERE
      cp.is_hidden = false
  AND ca.status = 'approved'
  AND (
    (COALESCE(@require_tags::TEXT[], '{}') = '{}' OR @require_tags::TEXT[] <@ cp.tags) OR
    (COALESCE(@include_hashtags::TEXT[], '{}') = '{}' OR (@include_hashtags::TEXT[] && cp.hashtags))
  )
  AND (
    (COALESCE(@exclude_tags::TEXT[], '{}') = '{}' OR NOT (@exclude_tags::TEXT[] && cp.tags)) AND
    (COALESCE(@exclude_hashtags::TEXT[], '{}') = '{}' OR NOT (@exclude_hashtags::TEXT[] && cp.hashtags))
  )
  AND (sqlc.narg(has_media)::BOOLEAN IS NULL OR sqlc.narg(has_media)::BOOLEAN = COALESCE(cp.has_media, true))
  AND (cp.indexed_at < @cursor_timestamp)
  AND cp.deleted_at IS NULL
ORDER BY
    cp.indexed_at DESC
LIMIT @_limit;

-- name: GetPostsWithLikes :many
SELECT
    cp.*,
    (SELECT
         COUNT(*)
     FROM
         candidate_likes cl
     WHERE
           cl.subject_uri = cp.uri
       AND (cl.indexed_at < @cursor_timestamp)
       AND cl.deleted_at IS NULL) AS likes
FROM
    candidate_posts cp
        INNER JOIN candidate_actors ca ON cp.actor_did = ca.did
WHERE
      cp.is_hidden = false
  AND ca.status = 'approved'
  AND (@cursor_timestamp::TIMESTAMPTZ IS NULL OR
       cp.indexed_at < @cursor_timestamp)
  AND cp.deleted_at IS NULL
ORDER BY
    cp.indexed_at DESC
LIMIT @_limit;
