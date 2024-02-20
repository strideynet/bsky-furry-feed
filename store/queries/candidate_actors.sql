-- name: ListCandidateActors :many
SELECT *
FROM
    candidate_actors AS ca
WHERE
    ca.status != 'none'
    AND (
        sqlc.narg(status)::actor_status IS NULL
        OR ca.status = sqlc.narg(status)
    )
ORDER BY
    ca.did;

-- name: CreateCandidateActor :one
INSERT INTO
candidate_actors (did, created_at, is_artist, comment, status, roles)
VALUES
($1, $2, $3, $4, $5, $6)
ON CONFLICT (did) DO UPDATE SET is_artist = excluded.is_artist,
comment = excluded.comment,
status = excluded.status,
roles = excluded.roles
RETURNING *;

-- name: UpdateCandidateActor :one
UPDATE candidate_actors ca
SET
    status = COALESCE(sqlc.narg(status), ca.status),
    is_artist = COALESCE(sqlc.narg(is_artist), ca.is_artist),
    comment = COALESCE(sqlc.narg(comment), ca.comment)
WHERE
    did = sqlc.arg(did)
RETURNING *;

-- name: CreateLatestActorProfile :exec
WITH
ap AS (
    INSERT INTO actor_profiles
    (
        actor_did, commit_cid, created_at, indexed_at, display_name,
        description, self_labels
    )
    VALUES
    (
        sqlc.arg(actor_did), sqlc.arg(commit_cid),
        sqlc.arg(created_at), sqlc.arg(indexed_at),
        sqlc.arg(display_name), sqlc.arg(description),
        sqlc.arg(self_labels)
    )
    ON CONFLICT (actor_did, commit_cid) DO
    UPDATE SET
    created_at = excluded.created_at,
    indexed_at = excluded.indexed_at,
    display_name = excluded.display_name,
    description = excluded.description,
    self_labels = excluded.self_labels
    RETURNING actor_did, commit_cid
)

UPDATE candidate_actors ca
SET
    current_profile_commit_cid = (SELECT commit_cid FROM ap)
WHERE
    did = (SELECT actor_did FROM ap);

-- name: GetLatestActorProfile :one
SELECT ap.*
FROM
    candidate_actors AS ca
INNER JOIN actor_profiles AS ap
    ON
        ca.did = ap.actor_did
        AND ca.current_profile_commit_cid
        = ap.commit_cid
WHERE
    ca.did = $1;

-- name: GetActorProfileHistory :many
SELECT ap.*
FROM
    actor_profiles AS ap
WHERE
    ap.actor_did = $1
ORDER BY
    ap.created_at DESC;

-- name: GetCandidateActorByDID :one
SELECT *
FROM
    candidate_actors
WHERE
    did = $1;

-- name: ListCandidateActorsRequiringProfileBackfill :many
SELECT *
FROM
    candidate_actors AS ca
WHERE
    ca.status = 'approved'
    AND ca.current_profile_commit_cid IS NULL
ORDER BY
    ca.did;

-- name: HoldBackPendingActor :exec
UPDATE candidate_actors ca
SET
    held_until = $1
WHERE
    ca.status = 'pending'
    AND ca.did = sqlc.arg(did);

-- name: OptInOrMarkActorPending :one
UPDATE candidate_actors ca
SET
    status
    = CASE WHEN ca.status = 'opted_out' THEN 'approved' WHEN ca.status = 'none' THEN 'pending' ELSE ca.status END
WHERE
    ca.did = sqlc.arg(did)
RETURNING status;

-- name: OptOutOrForgetActor :one
UPDATE candidate_actors ca
SET
    status
    = CASE WHEN ca.status = 'approved' THEN 'opted_out' WHEN ca.status = 'pending' THEN 'none' ELSE ca.status END
WHERE
    ca.did = sqlc.arg(did)
RETURNING status;
