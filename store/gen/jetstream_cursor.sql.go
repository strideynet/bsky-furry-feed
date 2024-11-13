// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: jetstream_cursor.sql

package gen

import (
	"context"
)

const getJetstreamCursor = `-- name: GetJetstreamCursor :one
SELECT cursor FROM jetstream_cursor
`

func (q *Queries) GetJetstreamCursor(ctx context.Context) (int64, error) {
	row := q.db.QueryRow(ctx, getJetstreamCursor)
	var cursor int64
	err := row.Scan(&cursor)
	return cursor, err
}

const setJetstreamCursor = `-- name: SetJetstreamCursor :exec
INSERT INTO jetstream_cursor (cursor)
VALUES ($1)
ON CONFLICT ((0)) DO
UPDATE SET cursor = excluded.cursor
`

func (q *Queries) SetJetstreamCursor(ctx context.Context, cursor int64) error {
	_, err := q.db.Exec(ctx, setJetstreamCursor, cursor)
	return err
}