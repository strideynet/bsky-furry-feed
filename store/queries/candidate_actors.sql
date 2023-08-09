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
    candidate_actors (did, created_at, is_artist, comment, status, roles)
VALUES
    ($1, $2, $3, $4, $5, $6)
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

-- name: CreateLatestActorProfile :exec
WITH
    ap as (
        INSERT INTO actor_profiles
            (actor_did, id, created_at, indexed_at, display_name, description,
             self_labels)
            VALUES
                (sqlc.arg(did), sqlc.arg(id), sqlc.arg(created_at),
                 sqlc.arg(indexed_at), sqlc.arg(display_name),
                 sqlc.arg(description), sqlc.arg(self_labels))
            RETURNING actor_did, id)
UPDATE candidate_actors ca
SET
    current_profile_id = (SELECT id FROM ap)
WHERE
    did = (SELECT actor_did FROM ap);

-- name: GetCandidateActorByDID :one
SELECT *
FROM
    candidate_actors
WHERE
    did = $1;

-- name: ListCandidateActorsRequiringProfileBackfill :many
SELECT *
FROM
    candidate_actors ca
WHERE
      ca.status = 'approved'
  AND ca.current_profile_id IS NULL
ORDER BY
    did;
