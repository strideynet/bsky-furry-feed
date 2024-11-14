-- name: SetJetstreamCursor :exec
INSERT INTO jetstream_cursor (cursor)
VALUES ($1)
ON CONFLICT ((0)) DO
UPDATE SET cursor = excluded.cursor;

-- name: GetJetstreamCursor :one
SELECT cursor FROM jetstream_cursor;
