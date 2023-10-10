// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.19.0
// source: candidate_actors.sql

package gen

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createCandidateActor = `-- name: CreateCandidateActor :one
INSERT INTO
candidate_actors (did, created_at, is_artist, comment, status, roles)
VALUES
($1, $2, $3, $4, $5, $6)
RETURNING did, created_at, is_artist, comment, status, roles, current_profile_commit_cid, held_until
`

type CreateCandidateActorParams struct {
	DID       string
	CreatedAt pgtype.Timestamptz
	IsArtist  bool
	Comment   string
	Status    ActorStatus
	Roles     []string
}

func (q *Queries) CreateCandidateActor(ctx context.Context, arg CreateCandidateActorParams) (CandidateActor, error) {
	row := q.db.QueryRow(ctx, createCandidateActor,
		arg.DID,
		arg.CreatedAt,
		arg.IsArtist,
		arg.Comment,
		arg.Status,
		arg.Roles,
	)
	var i CandidateActor
	err := row.Scan(
		&i.DID,
		&i.CreatedAt,
		&i.IsArtist,
		&i.Comment,
		&i.Status,
		&i.Roles,
		&i.CurrentProfileCommitCid,
		&i.HeldUntil,
	)
	return i, err
}

const createLatestActorProfile = `-- name: CreateLatestActorProfile :exec
WITH
ap AS (
    INSERT INTO actor_profiles
    (
        actor_did, commit_cid, created_at, indexed_at, display_name,
        description, self_labels
    )
    VALUES
    (
        $1, $2,
        $3, $4,
        $5, $6,
        $7
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
    did = (SELECT actor_did FROM ap)
`

type CreateLatestActorProfileParams struct {
	ActorDID    string
	CommitCID   string
	CreatedAt   pgtype.Timestamptz
	IndexedAt   pgtype.Timestamptz
	DisplayName pgtype.Text
	Description pgtype.Text
	SelfLabels  []string
}

func (q *Queries) CreateLatestActorProfile(ctx context.Context, arg CreateLatestActorProfileParams) error {
	_, err := q.db.Exec(ctx, createLatestActorProfile,
		arg.ActorDID,
		arg.CommitCID,
		arg.CreatedAt,
		arg.IndexedAt,
		arg.DisplayName,
		arg.Description,
		arg.SelfLabels,
	)
	return err
}

const getActorProfileHistory = `-- name: GetActorProfileHistory :many
SELECT ap.actor_did, ap.commit_cid, ap.created_at, ap.indexed_at, ap.display_name, ap.description, ap.self_labels
FROM
    actor_profiles AS ap
WHERE
    ap.actor_did = $1
ORDER BY
    created_at DESC
`

func (q *Queries) GetActorProfileHistory(ctx context.Context, actorDid string) ([]ActorProfile, error) {
	rows, err := q.db.Query(ctx, getActorProfileHistory, actorDid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ActorProfile
	for rows.Next() {
		var i ActorProfile
		if err := rows.Scan(
			&i.ActorDID,
			&i.CommitCID,
			&i.CreatedAt,
			&i.IndexedAt,
			&i.DisplayName,
			&i.Description,
			&i.SelfLabels,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getCandidateActorByDID = `-- name: GetCandidateActorByDID :one
SELECT did, created_at, is_artist, comment, status, roles, current_profile_commit_cid, held_until
FROM
    candidate_actors
WHERE
    did = $1
`

func (q *Queries) GetCandidateActorByDID(ctx context.Context, did string) (CandidateActor, error) {
	row := q.db.QueryRow(ctx, getCandidateActorByDID, did)
	var i CandidateActor
	err := row.Scan(
		&i.DID,
		&i.CreatedAt,
		&i.IsArtist,
		&i.Comment,
		&i.Status,
		&i.Roles,
		&i.CurrentProfileCommitCid,
		&i.HeldUntil,
	)
	return i, err
}

const getLatestActorProfile = `-- name: GetLatestActorProfile :one
SELECT ap.actor_did, ap.commit_cid, ap.created_at, ap.indexed_at, ap.display_name, ap.description, ap.self_labels
FROM
    candidate_actors AS ca
INNER JOIN actor_profiles AS ap
    ON
        ca.did = ap.actor_did
        AND ca.current_profile_commit_cid
        = ap.commit_cid
WHERE
    ca.did = $1
`

func (q *Queries) GetLatestActorProfile(ctx context.Context, did string) (ActorProfile, error) {
	row := q.db.QueryRow(ctx, getLatestActorProfile, did)
	var i ActorProfile
	err := row.Scan(
		&i.ActorDID,
		&i.CommitCID,
		&i.CreatedAt,
		&i.IndexedAt,
		&i.DisplayName,
		&i.Description,
		&i.SelfLabels,
	)
	return i, err
}

const holdBackPendingActor = `-- name: HoldBackPendingActor :exec
UPDATE candidate_actors ca
SET
    held_until = $1
WHERE
    ca.status = 'pending'
    AND ca.did = $2
`

type HoldBackPendingActorParams struct {
	HeldUntil pgtype.Timestamptz
	DID       string
}

func (q *Queries) HoldBackPendingActor(ctx context.Context, arg HoldBackPendingActorParams) error {
	_, err := q.db.Exec(ctx, holdBackPendingActor, arg.HeldUntil, arg.DID)
	return err
}

const listCandidateActors = `-- name: ListCandidateActors :many
SELECT did, created_at, is_artist, comment, status, roles, current_profile_commit_cid, held_until
FROM
    candidate_actors AS ca
WHERE
    (
        $1::actor_status IS NULL
        OR ca.status = $1
    )
ORDER BY
    did
`

func (q *Queries) ListCandidateActors(ctx context.Context, status NullActorStatus) ([]CandidateActor, error) {
	rows, err := q.db.Query(ctx, listCandidateActors, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []CandidateActor
	for rows.Next() {
		var i CandidateActor
		if err := rows.Scan(
			&i.DID,
			&i.CreatedAt,
			&i.IsArtist,
			&i.Comment,
			&i.Status,
			&i.Roles,
			&i.CurrentProfileCommitCid,
			&i.HeldUntil,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listCandidateActorsRequiringProfileBackfill = `-- name: ListCandidateActorsRequiringProfileBackfill :many
SELECT did, created_at, is_artist, comment, status, roles, current_profile_commit_cid, held_until
FROM
    candidate_actors AS ca
WHERE
    ca.status = 'approved'
    AND ca.current_profile_commit_cid IS NULL
ORDER BY
    did
`

func (q *Queries) ListCandidateActorsRequiringProfileBackfill(ctx context.Context) ([]CandidateActor, error) {
	rows, err := q.db.Query(ctx, listCandidateActorsRequiringProfileBackfill)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []CandidateActor
	for rows.Next() {
		var i CandidateActor
		if err := rows.Scan(
			&i.DID,
			&i.CreatedAt,
			&i.IsArtist,
			&i.Comment,
			&i.Status,
			&i.Roles,
			&i.CurrentProfileCommitCid,
			&i.HeldUntil,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateCandidateActor = `-- name: UpdateCandidateActor :one
UPDATE candidate_actors ca
SET
    status = COALESCE($1, ca.status),
    is_artist = COALESCE($2, ca.is_artist),
    comment = COALESCE($3, ca.comment)
WHERE
    did = $4
RETURNING did, created_at, is_artist, comment, status, roles, current_profile_commit_cid, held_until
`

type UpdateCandidateActorParams struct {
	Status   NullActorStatus
	IsArtist pgtype.Bool
	Comment  pgtype.Text
	DID      string
}

func (q *Queries) UpdateCandidateActor(ctx context.Context, arg UpdateCandidateActorParams) (CandidateActor, error) {
	row := q.db.QueryRow(ctx, updateCandidateActor,
		arg.Status,
		arg.IsArtist,
		arg.Comment,
		arg.DID,
	)
	var i CandidateActor
	err := row.Scan(
		&i.DID,
		&i.CreatedAt,
		&i.IsArtist,
		&i.Comment,
		&i.Status,
		&i.Roles,
		&i.CurrentProfileCommitCid,
		&i.HeldUntil,
	)
	return i, err
}
