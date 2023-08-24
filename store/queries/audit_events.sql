-- name: ListAuditEvents :many
SELECT *
FROM
    audit_events AS ae
WHERE
    (
        sqlc.arg(subject_did)::text = ''
        OR ae.subject_did = sqlc.arg(subject_did)
    )
    AND (
        sqlc.arg(actor_did)::text = ''
        OR ae.actor_did = sqlc.arg(actor_did)
    )
    AND (
        sqlc.arg(subject_record_uri)::text = ''
        OR ae.subject_record_uri = sqlc.arg(subject_record_uri)
    )
    AND (sqlc.arg(created_before)::timestamptz IS NULL OR ae.created_at < sqlc.arg(created_before))
ORDER BY
    ae.created_at DESC
LIMIT sqlc.arg(_limit);

-- name: CreateAuditEvent :one
INSERT INTO
audit_events (
    id, created_at, actor_did, subject_did, subject_record_uri,
    payload
)
VALUES
($1, $2, $3, $4, $5, $6)
RETURNING *;
