-- name: CreateCandidatePost :exec
INSERT INTO
    candidate_posts (uri, actor_did, created_at, indexed_at)
VALUES
    ($1, $2, $3, $4);

-- name: ListCandidatePostsForFeed :many
SELECT
    cp.*
FROM
    candidate_posts cp
        LEFT JOIN candidate_actors ca on cp.actor_did = ca.did
WHERE
      cp.is_hidden = false
  AND ca.is_hidden = false
ORDER BY
    cp.created_at DESC
LIMIT $1;

-- name: ListCandidatePostsForFeedWithCursor :many
SELECT
    cp.*
FROM
    candidate_posts cp
        LEFT JOIN candidate_actors ca on cp.actor_did = ca.did
WHERE
      cp.is_hidden = false
  AND ca.is_hidden = false
  AND cp.created_at < $1
ORDER BY
    cp.created_at DESC
LIMIT $2;

-- name: ListCandidatePostsForHotFeed :many
SELECT
    cp.*
FROM
    candidate_posts cp
        LEFT JOIN candidate_actors ca on cp.actor_did = ca.did
        left join candidate_likes cl on cp.uri = cl.subject_uri
WHERE
      cp.is_hidden = false
  AND ca.is_hidden = false
GROUP BY
    cp.uri
HAVING
    count(*) > 5
ORDER BY
    cp.created_at DESC
LIMIT $1;

-- name: ListCandidatePostsForHotFeedWithCursor :many
SELECT
    cp.*
FROM
    candidate_posts cp
        LEFT JOIN candidate_actors ca on cp.actor_did = ca.did
        left join candidate_likes cl on cp.uri = cl.subject_uri
WHERE
      cp.is_hidden = false
  AND ca.is_hidden = false
  AND cp.created_at < $1
GROUP BY
    cp.uri
HAVING
    count(*) > 5
ORDER BY
    cp.created_at DESC
LIMIT $2;


