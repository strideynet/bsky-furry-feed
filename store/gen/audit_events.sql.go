// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: audit_events.sql

package gen

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const CreateAuditEvent = `-- name: CreateAuditEvent :one
INSERT INTO
audit_events (
    id, created_at, actor_did, subject_did, subject_record_uri,
    payload
)
VALUES
($1, $2, $3, $4, $5, $6)
RETURNING id, actor_did, subject_did, subject_record_uri, created_at, payload
`

type CreateAuditEventParams struct {
	ID               string
	CreatedAt        pgtype.Timestamptz
	ActorDID         string
	SubjectDid       string
	SubjectRecordUri string
	Payload          []byte
}

func (q *Queries) CreateAuditEvent(ctx context.Context, arg CreateAuditEventParams) (AuditEvent, error) {
	row := q.db.QueryRow(ctx, CreateAuditEvent,
		arg.ID,
		arg.CreatedAt,
		arg.ActorDID,
		arg.SubjectDid,
		arg.SubjectRecordUri,
		arg.Payload,
	)
	var i AuditEvent
	err := row.Scan(
		&i.ID,
		&i.ActorDID,
		&i.SubjectDid,
		&i.SubjectRecordUri,
		&i.CreatedAt,
		&i.Payload,
	)
	return i, err
}

const ListAuditEvents = `-- name: ListAuditEvents :many
SELECT id, actor_did, subject_did, subject_record_uri, created_at, payload
FROM
    audit_events AS ae
WHERE
    (
        $1::text = ''
        OR ae.subject_did = $1
    )
    AND (
        $2::text = ''
        OR ae.actor_did = $2
    )
    AND (
        $3::text = ''
        OR ae.subject_record_uri = $3
    )
    AND ($4::timestamptz IS NULL OR ae.created_at < $4)
    AND (
        coalesce(cardinality($5::text []), 0) = 0
        OR (
            'COMMENT' = any($5)
            AND payload ->> '@type' = 'type.googleapis.com/bff.v1.CommentAuditPayload'
        )
        OR (
            'APPROVED' = any($5)
            AND payload ->> '@type' = 'type.googleapis.com/bff.v1.ProcessApprovalQueueAuditPayload'
            AND payload ->> 'action' = 'APPROVAL_QUEUE_ACTION_APPROVE'
        )
        OR (
            'REJECTED' = any($5)
            AND payload ->> '@type' = 'type.googleapis.com/bff.v1.ProcessApprovalQueueAuditPayload'
            AND payload ->> 'action' = 'APPROVAL_QUEUE_ACTION_REJECT'
        )
        OR (
            'HELD_BACK' = any($5)
            AND payload ->> '@type' = 'type.googleapis.com/bff.v1.HoldBackPendingActorAuditPayload'
        )
        OR (
            'FORCE_APPROVED' = any($5)
            AND payload ->> '@type' = 'type.googleapis.com/bff.v1.ForceApproveActorAuditPayload'
        )
        OR (
            'UNAPPROVED' = any($5)
            AND payload ->> '@type' = 'type.googleapis.com/bff.v1.UnapproveActorAuditPayload'
        )
        OR (
            'TRACKED' = any($5)
            AND payload ->> '@type' = 'type.googleapis.com/bff.v1.CreateActorAuditPayload'
        )
        OR (
            'BANNED' = any($5)
            AND payload ->> '@type' = 'type.googleapis.com/bff.v1.BanActorAuditPayload'
        )
        OR (
            'ASSIGNED_ROLES' = any($5)
            AND payload ->> '@type' = 'type.googleapis.com/bff.v1.AssignRolesAuditPayload'
        )
    )
ORDER BY
    ae.created_at DESC
LIMIT $6
`

type ListAuditEventsParams struct {
	SubjectDid       string
	ActorDID         string
	SubjectRecordUri string
	CreatedBefore    pgtype.Timestamptz
	Types            []string
	Limit            int32
}

func (q *Queries) ListAuditEvents(ctx context.Context, arg ListAuditEventsParams) ([]AuditEvent, error) {
	rows, err := q.db.Query(ctx, ListAuditEvents,
		arg.SubjectDid,
		arg.ActorDID,
		arg.SubjectRecordUri,
		arg.CreatedBefore,
		arg.Types,
		arg.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []AuditEvent
	for rows.Next() {
		var i AuditEvent
		if err := rows.Scan(
			&i.ID,
			&i.ActorDID,
			&i.SubjectDid,
			&i.SubjectRecordUri,
			&i.CreatedAt,
			&i.Payload,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
