-- name: CreateCandidatePost :exec
INSERT INTO
    candidate_posts (uri, actor_did, created_at, indexed_at, tags)
VALUES
    ($1, $2, $3, $4, $5);

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
  AND (@tag::TEXT = '' OR @tag::TEXT = ANY (cp.tags))
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

