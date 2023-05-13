-- name: CreateCandidatePost :exec
INSERT INTO candidate_posts (
    uri, repository_did, created_at, indexed_at
) VALUES (
    $1, $2, $3, $4
 );
