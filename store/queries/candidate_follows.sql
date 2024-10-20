-- name: CreateCandidateFollow :exec
INSERT INTO
candidate_follows (
    uri, actor_did, subject_did, created_at,
    indexed_at
)
VALUES
($1, $2, $3, $4, $5);

-- name: SoftDeleteCandidateFollow :one
UPDATE
candidate_follows
SET
    deleted_at = NOW()
WHERE
    uri = $1
RETURNING subject_did;
