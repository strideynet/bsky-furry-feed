-- name: SetFirehoseCommitCursor :exec
INSERT INTO firehose_commit_cursor (cursor)
VALUES ($1)
ON CONFLICT ((0)) DO
UPDATE SET cursor = excluded.cursor;

-- name: GetFirehoseCommitCursor :one
SELECT cursor FROM firehose_commit_cursor;
