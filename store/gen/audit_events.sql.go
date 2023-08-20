// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.19.0
// source: audit_events.sql

package gen

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createAuditEvent = `-- name: CreateAuditEvent :one
INSERT INTO
    audit_events (id, created_at, actor_did, subject_did, subject_record_uri,
                  payload)
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
	row := q.db.QueryRow(ctx, createAuditEvent,
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

const listAuditEvents = `-- name: ListAuditEvents :many
SELECT id, actor_did, subject_did, subject_record_uri, created_at, payload
FROM
    audit_events ae
WHERE
      ($1::text = '' OR
       ae.subject_did = $1)
  AND ($2::text = '' OR
       ae.actor_did = $2)
  AND ($3::text = '' OR
       ae.subject_record_uri = $3)
  AND ($4::TIMESTAMPTZ IS NULL OR ae.created_at < $4)
ORDER BY
    ae.created_at DESC
LIMIT $5
`

type ListAuditEventsParams struct {
	SubjectDid       string
	ActorDID         string
	SubjectRecordUri string
	CreatedBefore    pgtype.Timestamptz
	Limit            int32
}

func (q *Queries) ListAuditEvents(ctx context.Context, arg ListAuditEventsParams) ([]AuditEvent, error) {
	rows, err := q.db.Query(ctx, listAuditEvents,
		arg.SubjectDid,
		arg.ActorDID,
		arg.SubjectRecordUri,
		arg.CreatedBefore,
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
