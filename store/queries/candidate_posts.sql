-- name: CreateCandidatePost :exec
INSERT INTO
    candidate_posts (uri, actor_did, created_at, indexed_at)
VALUES
    ($1, $2, $3, $4);

-- name: GetFurryNewFeed :many
SELECT
    cp.*
FROM
    candidate_posts cp
        INNER JOIN candidate_actors ca ON cp.actor_did = ca.did
WHERE
      cp.is_hidden = false
  AND ca.status = 'approved'
  AND (@cursor_timestamp::TIMESTAMPTZ IS NULL OR
       cp.created_at < @cursor_timestamp)
ORDER BY
    cp.created_at DESC
LIMIT @_limit;

-- name: GetFurryHotFeed :many
SELECT
    cp.*
FROM
    candidate_posts cp
        INNER JOIN candidate_actors ca ON cp.actor_did = ca.did
        INNER JOIN candidate_likes cl ON cp.uri = cl.subject_uri
WHERE
      cp.is_hidden = false
  AND ca.status = 'approved'
  AND (@cursor_timestamp::TIMESTAMPTZ IS NULL OR
       cp.created_at < @cursor_timestamp)
GROUP BY
    cp.uri
HAVING
    count(*) >= @like_threshold::int
ORDER BY
    cp.created_at DESC
LIMIT @_limit;

