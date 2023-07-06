-- name: CreateCandidateLike :exec
INSERT INTO
    candidate_likes (uri, actor_did, subject_uri, created_at,
                     indexed_at)
VALUES
    ($1, $2, $3, $4, $5);

-- name: SoftDeleteCandidateLike :exec
UPDATE
    candidate_likes
SET
    deleted_at = NOW()
WHERE
    uri = $1;
