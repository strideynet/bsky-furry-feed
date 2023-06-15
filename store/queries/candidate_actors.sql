-- name: ListCandidateActors :many
SELECT *
FROM
    candidate_actors ca
WHERE
    (sqlc.narg(status)::actor_status IS NULL OR
     ca.status = @status)
ORDER BY
    did;

-- name: CreateCandidateActor :exec
INSERT INTO
    candidate_actors (did, created_at, is_artist, comment, status)
VALUES
    ($1, $2, $3, $4, $5);

-- name: UpdateCandidateActor :exec
UPDATE candidate_actors ca
SET
    status=COALESCE(sqlc.narg(status), ca.is_artist),
    is_artist=COALESCE(sqlc.narg(is_artist), ca.is_artist)
WHERE
    did = sqlc.arg(did);


-- name: GetCandidateActorByDID :one
SELECT *
FROM
    candidate_actors
WHERE
    did = $1;