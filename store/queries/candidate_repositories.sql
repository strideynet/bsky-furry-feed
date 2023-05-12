-- name: ListCandidateRepositories :many
SELECT * FROM candidate_repositories
ORDER BY did;

-- name: CreateCandidateRepository :one
INSERT INTO candidate_repositories (
    did, created_at, is_artist, comment
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;