-- name: GetCon :one
SELECT * FROM cons WHERE id = $1;
