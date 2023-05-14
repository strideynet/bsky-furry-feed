-- name: ListCandidateRepositories :many
SELECT * FROM candidate_repositories
ORDER BY did;

-- name: CreateCandidateRepository :exec
INSERT INTO candidate_repositories (
    did, created_at, is_artist, comment
) VALUES (
    $1, $2, $3, $4
);

-- name: GetCandidateRepositoryByDID :one
SELECT * FROM candidate_repositories WHERE did = $1;