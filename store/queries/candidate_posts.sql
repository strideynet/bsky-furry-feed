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
        LEFT JOIN candidate_actors ca ON cp.actor_did = ca.did
WHERE
      cp.is_hidden = false
  AND ca.is_hidden = false
  AND ('$1' IS NULL OR cp.created_at < $1)
ORDER BY
    cp.created_at DESC
LIMIT $2;

-- name: GetFurryHotFeed :many
SELECT
    cp.*
FROM
    candidate_posts cp
        INNER JOIN candidate_actors ca ON cp.actor_did = ca.did
        INNER JOIN candidate_likes cl ON cp.uri = cl.subject_uri
WHERE
      cp.is_hidden = false
  AND ca.is_hidden = false
  AND ('$1' IS NULL OR cp.created_at < $1)
GROUP BY
    cp.uri
HAVING
    count(*) > 4
ORDER BY
    cp.created_at DESC
LIMIT $2;

