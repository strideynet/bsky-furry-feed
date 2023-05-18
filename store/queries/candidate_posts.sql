-- name: CreateCandidatePost :exec
INSERT INTO candidate_posts (
    uri, actor_did, created_at, indexed_at
) VALUES (
    $1, $2, $3, $4
 );

-- name: ListCandidatePostsForFeed :many
SELECT * FROM candidate_posts ORDER BY created_at DESC LIMIT $1;

-- name: ListCandidatePostsForFeedWithCursor :many
SELECT * FROM candidate_posts WHERE created_at < $1 ORDER BY created_at DESC LIMIT $2;