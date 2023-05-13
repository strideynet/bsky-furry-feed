// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0
// source: candidate_posts.sql

package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createCandidatePost = `-- name: CreateCandidatePost :exec
INSERT INTO candidate_posts (
    uri, repository_did, created_at, indexed_at
) VALUES (
    $1, $2, $3, $4
 )
`

type CreateCandidatePostParams struct {
	URI           string
	RepositoryDID string
	CreatedAt     pgtype.Timestamptz
	IndexedAt     pgtype.Timestamptz
}

func (q *Queries) CreateCandidatePost(ctx context.Context, arg CreateCandidatePostParams) error {
	_, err := q.db.Exec(ctx, createCandidatePost,
		arg.URI,
		arg.RepositoryDID,
		arg.CreatedAt,
		arg.IndexedAt,
	)
	return err
}
