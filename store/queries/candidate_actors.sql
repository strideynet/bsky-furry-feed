-- name: ListCandidateActors :many
SELECT *
FROM
    candidate_actors
ORDER BY
    did;

-- name: CreateCandidateActor :exec
INSERT INTO
    candidate_actors (did, created_at, is_artist, comment)
VALUES
    ($1, $2, $3, $4);

-- name: GetCandidateActorByDID :one
SELECT *
FROM
    candidate_actors
WHERE
    did = $1;