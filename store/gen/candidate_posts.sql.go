// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0
// source: candidate_posts.sql

package gen

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createCandidatePost = `-- name: CreateCandidatePost :exec
INSERT INTO
    candidate_posts (uri, actor_did, created_at, indexed_at, tags)
VALUES
    ($1, $2, $3, $4, $5)
`

type CreateCandidatePostParams struct {
	URI       string
	ActorDID  string
	CreatedAt pgtype.Timestamptz
	IndexedAt pgtype.Timestamptz
	Tags      []string
}

func (q *Queries) CreateCandidatePost(ctx context.Context, db DBTX, arg CreateCandidatePostParams) error {
	_, err := db.Exec(ctx, createCandidatePost,
		arg.URI,
		arg.ActorDID,
		arg.CreatedAt,
		arg.IndexedAt,
		arg.Tags,
	)
	return err
}

const getFurryNewFeed = `-- name: GetFurryNewFeed :many
SELECT
    cp.uri, cp.actor_did, cp.created_at, cp.indexed_at, cp.is_nsfw, cp.is_hidden, cp.tags, cp.deleted_at
FROM
    candidate_posts cp
        INNER JOIN candidate_actors ca ON cp.actor_did = ca.did
WHERE
      cp.is_hidden = false
  AND ca.status = 'approved'
  AND ($1::TEXT = '' OR $1::TEXT = ANY (cp.tags))
  AND (cp.indexed_at < $2)
  AND cp.deleted_at IS NULL
ORDER BY
    cp.indexed_at DESC
LIMIT $3
`

type GetFurryNewFeedParams struct {
	Tag             string
	CursorTimestamp pgtype.Timestamptz
	Limit           int32
}

func (q *Queries) GetFurryNewFeed(ctx context.Context, db DBTX, arg GetFurryNewFeedParams) ([]CandidatePost, error) {
	rows, err := db.Query(ctx, getFurryNewFeed, arg.Tag, arg.CursorTimestamp, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []CandidatePost
	for rows.Next() {
		var i CandidatePost
		if err := rows.Scan(
			&i.URI,
			&i.ActorDID,
			&i.CreatedAt,
			&i.IndexedAt,
			&i.IsNSFW,
			&i.IsHidden,
			&i.Tags,
			&i.DeletedAt,
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

const getPostsWithLikes = `-- name: GetPostsWithLikes :many
SELECT
    cp.uri, cp.actor_did, cp.created_at, cp.indexed_at, cp.is_nsfw, cp.is_hidden, cp.tags, cp.deleted_at,
    (SELECT
         COUNT(*)
     FROM
         candidate_likes cl
     WHERE
           cl.subject_uri = cp.uri
       AND (cl.indexed_at < $1)
       AND cl.deleted_at IS NULL) AS likes
FROM
    candidate_posts cp
        INNER JOIN candidate_actors ca ON cp.actor_did = ca.did
WHERE
      cp.is_hidden = false
  AND ca.status = 'approved'
  AND ($1::TIMESTAMPTZ IS NULL OR
       cp.indexed_at < $1)
  AND cp.deleted_at IS NULL
ORDER BY
    cp.indexed_at DESC
LIMIT $2
`

type GetPostsWithLikesParams struct {
	CursorTimestamp pgtype.Timestamptz
	Limit           int32
}

type GetPostsWithLikesRow struct {
	URI       string
	ActorDID  string
	CreatedAt pgtype.Timestamptz
	IndexedAt pgtype.Timestamptz
	IsNSFW    bool
	IsHidden  bool
	Tags      []string
	DeletedAt pgtype.Timestamptz
	Likes     int64
}

func (q *Queries) GetPostsWithLikes(ctx context.Context, db DBTX, arg GetPostsWithLikesParams) ([]GetPostsWithLikesRow, error) {
	rows, err := db.Query(ctx, getPostsWithLikes, arg.CursorTimestamp, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetPostsWithLikesRow
	for rows.Next() {
		var i GetPostsWithLikesRow
		if err := rows.Scan(
			&i.URI,
			&i.ActorDID,
			&i.CreatedAt,
			&i.IndexedAt,
			&i.IsNSFW,
			&i.IsHidden,
			&i.Tags,
			&i.DeletedAt,
			&i.Likes,
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

const softDeleteCandidatePost = `-- name: SoftDeleteCandidatePost :exec
UPDATE
    candidate_posts
SET
    deleted_at = NOW()
WHERE
    uri = $1
`

func (q *Queries) SoftDeleteCandidatePost(ctx context.Context, db DBTX, uri string) error {
	_, err := db.Exec(ctx, softDeleteCandidatePost, uri)
	return err
}
