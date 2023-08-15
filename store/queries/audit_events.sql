-- name: ListAuditEvents :many
SELECT *
FROM
    audit_events ae
WHERE
      (@subject_did::text = '' OR
       ae.subject_did = @subject_did)
  AND (@actor_did::text = '' OR
       ae.actor_did = @actor_did)
  AND (@subject_record_uri::text = '' OR
       ae.subject_record_uri = @subject_record_uri)
  AND (@created_before::TIMESTAMPTZ IS NULL OR ae.created_at < @created_before)
ORDER BY
    ae.created_at DESC
LIMIT @_limit;

-- name: CreateAuditEvent :one
INSERT INTO
    audit_events (id, created_at, actor_did, subject_did, subject_record_uri,
                  payload)
VALUES
    ($1, $2, $3, $4, $5, $6)
RETURNING *;