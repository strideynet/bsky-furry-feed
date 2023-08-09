-- name: CreateCandidatePost :exec
INSERT INTO
    candidate_posts (uri, actor_did, created_at, indexed_at, hashtags,
                     has_media, raw, self_labels)
VALUES
    ($1, $2, $3, $4, $5, $6, $7, $8);

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
      -- Only include posts by approved actors
      ca.status = 'approved'
      -- Remove posts hidden by our moderators
  AND cp.is_hidden = false
      -- Remove posts deleted by the actors
  AND cp.deleted_at IS NULL
      -- Match at least one of the queried hashtags. If unspecified, do not filter.
  AND (COALESCE(@hashtags::TEXT[], '{}') = '{}' OR
       @hashtags::TEXT[] && cp.hashtags)
      -- Match has_media status. If unspecified, do not filter.
  AND (sqlc.narg(has_media)::BOOLEAN IS NULL OR
       COALESCE(cp.has_media, false) = @has_media)
      -- Filter by NSFW status. If unspecified, do not filter.
  AND (sqlc.narg(is_nsfw)::BOOLEAN IS NULL OR
       ((ARRAY ['nsfw', 'mursuit', 'murrsuit'] && cp.hashtags) OR
        ARRAY ['nsfw', 'mursuit', 'murrsuit'] && cp.self_labels) = @is_nsfw)

      -- Remove posts newer than the cursor timestamp
  AND (cp.indexed_at < @cursor_timestamp)
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

-- name: GetPostByURI :one
SELECT *
FROM
    candidate_posts cp
WHERE
    cp.uri = @uri
LIMIT 1;
