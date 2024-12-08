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
    AND (
        coalesce(cardinality(sqlc.arg(types)::text []), 0) = 0
        OR (
            'COMMENT' = any(sqlc.arg(types))
            AND payload ->> '@type' = 'type.googleapis.com/bff.v1.CommentAuditPayload'
        )
        OR (
            'APPROVED' = any(sqlc.arg(types))
            AND payload ->> '@type' = 'type.googleapis.com/bff.v1.ProcessApprovalQueueAuditPayload'
            AND payload ->> 'action' = 'APPROVAL_QUEUE_ACTION_APPROVE'
        )
        OR (
            'REJECTED' = any(sqlc.arg(types))
            AND payload ->> '@type' = 'type.googleapis.com/bff.v1.ProcessApprovalQueueAuditPayload'
            AND payload ->> 'action' = 'APPROVAL_QUEUE_ACTION_REJECT'
        )
        OR (
            'HELD_BACK' = any(sqlc.arg(types))
            AND payload ->> '@type' = 'type.googleapis.com/bff.v1.HoldBackPendingActorAuditPayload'
        )
        OR (
            'FORCE_APPROVED' = any(sqlc.arg(types))
            AND payload ->> '@type' = 'type.googleapis.com/bff.v1.ForceApproveActorAuditPayload'
        )
        OR (
            'UNAPPROVED' = any(sqlc.arg(types))
            AND payload ->> '@type' = 'type.googleapis.com/bff.v1.UnapproveActorAuditPayload'
        )
        OR (
            'TRACKED' = any(sqlc.arg(types))
            AND payload ->> '@type' = 'type.googleapis.com/bff.v1.CreateActorAuditPayload'
        )
        OR (
            'BANNED' = any(sqlc.arg(types))
            AND payload ->> '@type' = 'type.googleapis.com/bff.v1.BanActorAuditPayload'
        )
        OR (
            'ASSIGNED_ROLES' = any(sqlc.arg(types))
            AND payload ->> '@type' = 'type.googleapis.com/bff.v1.AssignRolesAuditPayload'
        )
    )
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
