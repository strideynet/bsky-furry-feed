// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.19.0
// source: candidate_posts.sql

package gen

import (
	"context"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/jackc/pgx/v5/pgtype"
)

const createCandidatePost = `-- name: CreateCandidatePost :exec
INSERT INTO
    candidate_posts (uri, actor_did, created_at, indexed_at, hashtags,
                     has_media, raw, self_labels)
VALUES
    ($1, $2, $3, $4, $5, $6, $7, $8)
`

type CreateCandidatePostParams struct {
	URI        string
	ActorDID   string
	CreatedAt  pgtype.Timestamptz
	IndexedAt  pgtype.Timestamptz
	Hashtags   []string
	HasMedia   pgtype.Bool
	Raw        *bsky.FeedPost
	SelfLabels []string
}

func (q *Queries) CreateCandidatePost(ctx context.Context, db DBTX, arg CreateCandidatePostParams) error {
	_, err := db.Exec(ctx, createCandidatePost,
		arg.URI,
		arg.ActorDID,
		arg.CreatedAt,
		arg.IndexedAt,
		arg.Hashtags,
		arg.HasMedia,
		arg.Raw,
		arg.SelfLabels,
	)
	return err
}

const getFurryNewFeed = `-- name: GetFurryNewFeed :many
SELECT
    cp.uri, cp.actor_did, cp.created_at, cp.indexed_at, cp.is_hidden, cp.deleted_at, cp.raw, cp.hashtags, cp.has_media, cp.self_labels
FROM
    candidate_posts cp
        INNER JOIN candidate_actors ca ON cp.actor_did = ca.did
WHERE
      -- Only include posts by approved actors
      ca.status = 'approved'
      -- Remove posts hidden by our moderators
  AND cp.is_hidden = false
      -- Remove posts deleted by the actors
  AND cp.deleted_at IS NULL
      -- Match at least one of the queried hashtags. If unspecified, do not filter.
  AND (COALESCE($1::TEXT[], '{}') = '{}' OR
       $1::TEXT[] && cp.hashtags)
      -- Match has_media status. If unspecified, do not filter.
  AND ($2::BOOLEAN IS NULL OR
       COALESCE(cp.has_media, false) = $2)
      -- Filter by NSFW status. If unspecified, do not filter.
  AND ($3::BOOLEAN IS NULL OR
       ((ARRAY ['nsfw', 'mursuit', 'murrsuit'] && cp.hashtags) OR
        ARRAY ['TODO-TEMPORARY'] && cp.self_labels) = $3)

      -- Remove posts newer than the cursor timestamp
  AND (cp.indexed_at < $4)
ORDER BY
    cp.indexed_at DESC
LIMIT $5
`

type GetFurryNewFeedParams struct {
	Hashtags        []string
	HasMedia        pgtype.Bool
	IsNSFW          pgtype.Bool
	CursorTimestamp pgtype.Timestamptz
	Limit           int32
}

func (q *Queries) GetFurryNewFeed(ctx context.Context, db DBTX, arg GetFurryNewFeedParams) ([]CandidatePost, error) {
	rows, err := db.Query(ctx, getFurryNewFeed,
		arg.Hashtags,
		arg.HasMedia,
		arg.IsNSFW,
		arg.CursorTimestamp,
		arg.Limit,
	)
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
			&i.IsHidden,
			&i.DeletedAt,
			&i.Raw,
			&i.Hashtags,
			&i.HasMedia,
			&i.SelfLabels,
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

const getPostByURI = `-- name: GetPostByURI :one
SELECT uri, actor_did, created_at, indexed_at, is_hidden, deleted_at, raw, hashtags, has_media, self_labels
FROM
    candidate_posts cp
WHERE
    cp.uri = $1
LIMIT 1
`

func (q *Queries) GetPostByURI(ctx context.Context, db DBTX, uri string) (CandidatePost, error) {
	row := db.QueryRow(ctx, getPostByURI, uri)
	var i CandidatePost
	err := row.Scan(
		&i.URI,
		&i.ActorDID,
		&i.CreatedAt,
		&i.IndexedAt,
		&i.IsHidden,
		&i.DeletedAt,
		&i.Raw,
		&i.Hashtags,
		&i.HasMedia,
		&i.SelfLabels,
	)
	return i, err
}

const getPostsWithLikes = `-- name: GetPostsWithLikes :many
SELECT
    cp.uri, cp.actor_did, cp.created_at, cp.indexed_at, cp.is_hidden, cp.deleted_at, cp.raw, cp.hashtags, cp.has_media, cp.self_labels,
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
	URI        string
	ActorDID   string
	CreatedAt  pgtype.Timestamptz
	IndexedAt  pgtype.Timestamptz
	IsHidden   bool
	DeletedAt  pgtype.Timestamptz
	Raw        *bsky.FeedPost
	Hashtags   []string
	HasMedia   pgtype.Bool
	SelfLabels []string
	Likes      int64
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
			&i.IsHidden,
			&i.DeletedAt,
			&i.Raw,
			&i.Hashtags,
			&i.HasMedia,
			&i.SelfLabels,
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
