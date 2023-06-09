-- name: ListCandidateActors :many
SELECT *
FROM
    candidate_actors ca
WHERE
    (sqlc.narg(status)::actor_status IS NULL OR
     ca.status = @status)
ORDER BY
    did;

-- name: CreateCandidateActor :one
INSERT INTO
    candidate_actors (did, created_at, is_artist, comment, status)
VALUES
    ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateCandidateActor :one
UPDATE candidate_actors ca
SET
    status=COALESCE(sqlc.narg(status), ca.status),
    is_artist=COALESCE(sqlc.narg(is_artist), ca.is_artist),
    comment=COALESCE(sqlc.narg(comment), ca.comment)
WHERE
    did = sqlc.arg(did)
RETURNING *;

-- name: GetCandidateActorByDID :one
SELECT *
FROM
    candidate_actors
WHERE
    did = $1;