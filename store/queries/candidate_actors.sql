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
            (actor_did, commit_cid, created_at, indexed_at, display_name,
             description)
            VALUES
                (sqlc.arg(actor_did), sqlc.arg(commit_cid),
                 sqlc.arg(created_at), sqlc.arg(indexed_at),
                 sqlc.arg(display_name), sqlc.arg(description))
            ON CONFLICT (actor_did, commit_cid) DO
                UPDATE SET
                    created_at = EXCLUDED.created_at,
                    indexed_at = EXCLUDED.indexed_at,
                    display_name = EXCLUDED.display_name,
                    description = EXCLUDED.description
            RETURNING actor_did, commit_cid)
UPDATE candidate_actors ca
SET
    current_profile_commit_cid = (SELECT commit_cid FROM ap)
WHERE
    did = (SELECT actor_did FROM ap);

-- name: GetLatestActorProfile :one
SELECT
    ap.*
FROM
    candidate_actors ca
        INNER JOIN actor_profiles ap ON ap.actor_did = ca.did AND
                                        ap.commit_cid =
                                        ca.current_profile_commit_cid
WHERE
    ca.did = $1;

-- name: GetActorProfileHistory :many
SELECT
    ap.*
FROM
    actor_profiles ap
WHERE
    ap.actor_did = $1
ORDER BY
    created_at DESC;

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
  AND ca.current_profile_commit_cid IS NULL
ORDER BY
    did;
