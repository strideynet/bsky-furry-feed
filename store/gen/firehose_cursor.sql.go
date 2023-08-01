// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.19.0
// source: firehose_cursor.sql

package gen

import (
	"context"
)

const getFirehoseCommitCursor = `-- name: GetFirehoseCommitCursor :one
SELECT cursor FROM firehose_commit_cursor
`

func (q *Queries) GetFirehoseCommitCursor(ctx context.Context, db DBTX) (int64, error) {
	row := db.QueryRow(ctx, getFirehoseCommitCursor)
	var cursor int64
	err := row.Scan(&cursor)
	return cursor, err
}

const setFirehoseCommitCursor = `-- name: SetFirehoseCommitCursor :exec
INSERT INTO firehose_commit_cursor (cursor)
VALUES ($1)
ON CONFLICT ((0)) DO
UPDATE SET cursor = EXCLUDED.cursor
`

func (q *Queries) SetFirehoseCommitCursor(ctx context.Context, db DBTX, cursor int64) error {
	_, err := db.Exec(ctx, setFirehoseCommitCursor, cursor)
	return err
}
