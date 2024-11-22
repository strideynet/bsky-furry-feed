// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: follow_tasks.sql

package gen

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const enqueueFollowTask = `-- name: EnqueueFollowTask :exec
INSERT INTO follow_tasks (
    actor_did,
    next_try_at,
    created_at,
    should_unfollow
) VALUES ($1, $2, $3, $4)
`

type EnqueueFollowTaskParams struct {
	ActorDID       string
	NextTryAt      pgtype.Timestamptz
	CreatedAt      pgtype.Timestamptz
	ShouldUnfollow bool
}

func (q *Queries) EnqueueFollowTask(ctx context.Context, arg EnqueueFollowTaskParams) error {
	_, err := q.db.Exec(ctx, enqueueFollowTask,
		arg.ActorDID,
		arg.NextTryAt,
		arg.CreatedAt,
		arg.ShouldUnfollow,
	)
	return err
}

const getNextFollowTask = `-- name: GetNextFollowTask :one
SELECT id, actor_did, next_try_at, created_at, tries, should_unfollow, finished_at, last_error
FROM follow_tasks
WHERE
    next_try_at <= NOW()
    AND finished_at IS NULL
    AND tries < 3
ORDER BY id ASC
LIMIT 1
`

func (q *Queries) GetNextFollowTask(ctx context.Context) (FollowTask, error) {
	row := q.db.QueryRow(ctx, getNextFollowTask)
	var i FollowTask
	err := row.Scan(
		&i.ID,
		&i.ActorDID,
		&i.NextTryAt,
		&i.CreatedAt,
		&i.Tries,
		&i.ShouldUnfollow,
		&i.FinishedAt,
		&i.LastError,
	)
	return i, err
}

const markFollowTaskAsDone = `-- name: MarkFollowTaskAsDone :exec
UPDATE follow_tasks
SET finished_at = NOW()
WHERE id = $1
`

func (q *Queries) MarkFollowTaskAsDone(ctx context.Context, id int64) error {
	_, err := q.db.Exec(ctx, markFollowTaskAsDone, id)
	return err
}

const markFollowTaskAsErrored = `-- name: MarkFollowTaskAsErrored :exec
UPDATE follow_tasks
SET
    next_try_at = $2,
    tries = tries + 1,
    last_error = $3
WHERE id = $1
`

type MarkFollowTaskAsErroredParams struct {
	ID        int64
	NextTryAt pgtype.Timestamptz
	LastError pgtype.Text
}

func (q *Queries) MarkFollowTaskAsErrored(ctx context.Context, arg MarkFollowTaskAsErroredParams) error {
	_, err := q.db.Exec(ctx, markFollowTaskAsErrored, arg.ID, arg.NextTryAt, arg.LastError)
	return err
}