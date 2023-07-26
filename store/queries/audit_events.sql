-- name: ListAuditEvents :many
SELECT *
FROM
    audit_events ae
WHERE
    (@subject_did::text  = '' OR
     ae.subject_did = @subject_did) AND
    (@actor_did::text  = '' OR
     ae.actor_did = @actor_did) AND
    (@subject_record_uri::text  = '' OR
     ae.subject_record_uri = @subject_record_uri)
ORDER BY
    ae.created_at DESC;
