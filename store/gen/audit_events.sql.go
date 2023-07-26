// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0
// source: audit_events.sql

package gen

import (
	"context"
)

const listAuditEvents = `-- name: ListAuditEvents :many
SELECT id, actor_did, subject_did, subject_record_uri, created_at, payload
FROM
    audit_events ae
WHERE
    ($1::text  = '' OR
     ae.subject_did = $1) AND
    ($2::text  = '' OR
     ae.actor_did = $2) AND
    ($3::text  = '' OR
     ae.subject_record_uri = $3)
ORDER BY
    ae.created_at DESC
`

type ListAuditEventsParams struct {
	SubjectDid       string
	ActorDID         string
	SubjectRecordUri string
}

func (q *Queries) ListAuditEvents(ctx context.Context, db DBTX, arg ListAuditEventsParams) ([]AuditEvent, error) {
	rows, err := db.Query(ctx, listAuditEvents, arg.SubjectDid, arg.ActorDID, arg.SubjectRecordUri)
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
