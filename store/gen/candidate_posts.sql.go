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
candidate_posts (
    uri, actor_did, created_at, indexed_at, hashtags, has_media, raw
)
VALUES
($1, $2, $3, $4, $5, $6, $7)
`

type CreateCandidatePostParams struct {
	URI       string
	ActorDID  string
	CreatedAt pgtype.Timestamptz
	IndexedAt pgtype.Timestamptz
	Hashtags  []string
	HasMedia  pgtype.Bool
	Raw       *bsky.FeedPost
}

func (q *Queries) CreateCandidatePost(ctx context.Context, arg CreateCandidatePostParams) error {
	_, err := q.db.Exec(ctx, createCandidatePost,
		arg.URI,
		arg.ActorDID,
		arg.CreatedAt,
		arg.IndexedAt,
		arg.Hashtags,
		arg.HasMedia,
		arg.Raw,
	)
	return err
}

const getFurryNewFeed = `-- name: GetFurryNewFeed :many
SELECT cp.uri, cp.actor_did, cp.created_at, cp.indexed_at, cp.is_hidden, cp.deleted_at, cp.raw, cp.hashtags, cp.has_media
FROM
    candidate_posts AS cp
INNER JOIN candidate_actors AS ca ON cp.actor_did = ca.did
WHERE
    cp.is_hidden = false
    AND ca.status = 'approved'
    AND (
        COALESCE($1::TEXT [], '{}') = '{}'
        OR $1::TEXT [] && cp.hashtags
    )
    AND (
        $2::BOOLEAN IS NULL
        OR COALESCE(cp.has_media, false) = $2
    )
    AND (
        $3::BOOLEAN IS NULL
        OR (ARRAY['nsfw', 'mursuit', 'murrsuit'] && cp.hashtags)
        = $3
    )
    AND (cp.indexed_at < $4)
    AND cp.deleted_at IS NULL
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

func (q *Queries) GetFurryNewFeed(ctx context.Context, arg GetFurryNewFeedParams) ([]CandidatePost, error) {
	rows, err := q.db.Query(ctx, getFurryNewFeed,
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

const getHotPosts = `-- name: GetHotPosts :many
SELECT
    cp.uri, cp.actor_did, cp.created_at, cp.indexed_at, cp.is_hidden, cp.deleted_at, cp.raw, cp.hashtags, cp.has_media,
    ph.score
FROM
    candidate_posts AS cp
INNER JOIN candidate_actors AS ca ON cp.actor_did = ca.did
INNER JOIN post_hotness AS ph
    ON
        ph.post_uri = cp.uri AND ph.alg = $1
        AND ph.generation_seq = $2
WHERE
    cp.is_hidden = false
    AND ca.status = 'approved'
    AND (
        COALESCE($3::TEXT [], '{}') = '{}'
        OR $3::TEXT [] && cp.hashtags
    )
    AND (
        $4::BOOLEAN IS NULL
        OR COALESCE(cp.has_media, false) = $4
    )
    AND (
        $5::BOOLEAN IS NULL
        OR (ARRAY['nsfw', 'mursuit', 'murrsuit'] && cp.hashtags)
        = $5
    )
    AND cp.deleted_at IS NULL
    AND (
        (ph.score, ph.uri)
        < ($6::REAL, $7::TEXT)
    )
ORDER BY
    ph.score DESC, ph.uri DESC
LIMIT $8
`

type GetHotPostsParams struct {
	Alg           string
	GenerationSeq int64
	Hashtags      []string
	HasMedia      pgtype.Bool
	IsNSFW        pgtype.Bool
	AfterScore    float32
	AfterURI      string
	Limit         int32
}

type GetHotPostsRow struct {
	URI       string
	ActorDID  string
	CreatedAt pgtype.Timestamptz
	IndexedAt pgtype.Timestamptz
	IsHidden  bool
	DeletedAt pgtype.Timestamptz
	Raw       *bsky.FeedPost
	Hashtags  []string
	HasMedia  pgtype.Bool
	Score     float32
}

func (q *Queries) GetHotPosts(ctx context.Context, arg GetHotPostsParams) ([]GetHotPostsRow, error) {
	rows, err := q.db.Query(ctx, getHotPosts,
		arg.Alg,
		arg.GenerationSeq,
		arg.Hashtags,
		arg.HasMedia,
		arg.IsNSFW,
		arg.AfterScore,
		arg.AfterURI,
		arg.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetHotPostsRow
	for rows.Next() {
		var i GetHotPostsRow
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
			&i.Score,
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

const getLatestHotPostGeneration = `-- name: GetLatestHotPostGeneration :one
SELECT ph.generation_seq
FROM post_hotness AS ph
WHERE ph.alg = $1
ORDER BY ph.generation_seq DESC
LIMIT 1
`

func (q *Queries) GetLatestHotPostGeneration(ctx context.Context, alg string) (int64, error) {
	row := q.db.QueryRow(ctx, getLatestHotPostGeneration, alg)
	var generation_seq int64
	err := row.Scan(&generation_seq)
	return generation_seq, err
}

const getPostByURI = `-- name: GetPostByURI :one
SELECT uri, actor_did, created_at, indexed_at, is_hidden, deleted_at, raw, hashtags, has_media
FROM
    candidate_posts AS cp
WHERE
    cp.uri = $1
LIMIT 1
`

func (q *Queries) GetPostByURI(ctx context.Context, uri string) (CandidatePost, error) {
	row := q.db.QueryRow(ctx, getPostByURI, uri)
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
	)
	return i, err
}

const getPostsWithLikes = `-- name: GetPostsWithLikes :many
SELECT
    cp.uri, cp.actor_did, cp.created_at, cp.indexed_at, cp.is_hidden, cp.deleted_at, cp.raw, cp.hashtags, cp.has_media,
    (
        SELECT COUNT(*)
        FROM
            candidate_likes AS cl
        WHERE
            cl.subject_uri = cp.uri
            AND (cl.indexed_at < $1)
            AND cl.deleted_at IS NULL
    ) AS likes
FROM
    candidate_posts AS cp
INNER JOIN candidate_actors AS ca ON cp.actor_did = ca.did
WHERE
    cp.is_hidden = false
    AND ca.status = 'approved'
    AND (
        $1::TIMESTAMPTZ IS NULL
        OR cp.indexed_at < $1
    )
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
	IsHidden  bool
	DeletedAt pgtype.Timestamptz
	Raw       *bsky.FeedPost
	Hashtags  []string
	HasMedia  pgtype.Bool
	Likes     int64
}

func (q *Queries) GetPostsWithLikes(ctx context.Context, arg GetPostsWithLikesParams) ([]GetPostsWithLikesRow, error) {
	rows, err := q.db.Query(ctx, getPostsWithLikes, arg.CursorTimestamp, arg.Limit)
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

func (q *Queries) SoftDeleteCandidatePost(ctx context.Context, uri string) error {
	_, err := q.db.Exec(ctx, softDeleteCandidatePost, uri)
	return err
}
