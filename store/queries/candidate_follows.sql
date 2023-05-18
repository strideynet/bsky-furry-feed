-- name: CreateCandidateFollow :exec
INSERT INTO
    candidate_follows (uri, actor_did, subject_did, created_at,
                     indexed_at)
VALUES
    ($1, $2, $3, $4, $5);