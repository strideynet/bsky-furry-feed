// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
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
    uri,
    actor_did,
    created_at,
    indexed_at,
    hashtags,
    has_media,
    has_video,
    raw,
    self_labels
)
VALUES
($1, $2, $3, $4, $5, $6, $7, $8, $9)
`

type CreateCandidatePostParams struct {
	URI        string
	ActorDID   string
	CreatedAt  pgtype.Timestamptz
	IndexedAt  pgtype.Timestamptz
	Hashtags   []string
	HasMedia   pgtype.Bool
	HasVideo   pgtype.Bool
	Raw        *bsky.FeedPost
	SelfLabels []string
}

func (q *Queries) CreateCandidatePost(ctx context.Context, arg CreateCandidatePostParams) error {
	_, err := q.db.Exec(ctx, createCandidatePost,
		arg.URI,
		arg.ActorDID,
		arg.CreatedAt,
		arg.IndexedAt,
		arg.Hashtags,
		arg.HasMedia,
		arg.HasVideo,
		arg.Raw,
		arg.SelfLabels,
	)
	return err
}

const getFurryNewFeed = `-- name: GetFurryNewFeed :many
SELECT cp.uri, cp.actor_did, cp.created_at, cp.indexed_at, cp.is_hidden, cp.deleted_at, cp.raw, cp.hashtags, cp.has_media, cp.self_labels, cp.has_video
FROM
    candidate_posts AS cp
INNER JOIN candidate_actors AS ca ON cp.actor_did = ca.did
WHERE
    -- Only include posts by approved actors
    ca.status = 'approved'
    -- Remove posts hidden by our moderators
    AND cp.is_hidden = FALSE
    -- Remove posts deleted by the actors
    AND cp.deleted_at IS NULL
    AND (
    -- Standard criteria.
        (
            -- Match at least one of the queried hashtags.
            -- If unspecified, do not filter.
            (
                COALESCE($1::TEXT [], '{}') = '{}'
                OR $1::TEXT [] && cp.hashtags
            )
            -- If any hashtags are disallowed, filter them out.
            AND (
                COALESCE($2::TEXT [], '{}') = '{}'
                OR NOT $2::TEXT [] && cp.hashtags
            )
            AND (
                -- Match has_media status. If unspecified, do not filter.
                (
                    $3::BOOLEAN IS NULL
                    OR COALESCE(cp.has_media, FALSE) = $3
                )
                -- Match has_video status. If unspecified, do not filter.
                OR (
                    $4::BOOLEAN IS NULL
                    OR COALESCE(cp.has_video, FALSE) = $4
                )
            )
            -- Filter by NSFW status. If unspecified, do not filter.
            AND (
                $5::BOOLEAN IS NULL
                OR (
                    (ARRAY['nsfw', 'mursuit', 'murrsuit', 'nsfwfurry', 'furrynsfw'] && cp.hashtags)
                    OR (ARRAY['porn', 'nudity', 'sexual'] && cp.self_labels)
                ) = $5
            )
        )
        -- Pinned DID criteria.
        OR cp.actor_did = ANY($6::TEXT [])
    )
    -- Remove posts newer than the cursor timestamp
    AND (cp.indexed_at < $7)
    AND cp.indexed_at > NOW() - INTERVAL '7 day'
ORDER BY
    cp.indexed_at DESC
LIMIT $8
`

type GetFurryNewFeedParams struct {
	Hashtags           []string
	DisallowedHashtags []string
	HasMedia           pgtype.Bool
	HasVideo           pgtype.Bool
	IsNSFW             pgtype.Bool
	PinnedDIDs         []string
	CursorTimestamp    pgtype.Timestamptz
	Limit              int32
}

func (q *Queries) GetFurryNewFeed(ctx context.Context, arg GetFurryNewFeedParams) ([]CandidatePost, error) {
	rows, err := q.db.Query(ctx, getFurryNewFeed,
		arg.Hashtags,
		arg.DisallowedHashtags,
		arg.HasMedia,
		arg.HasVideo,
		arg.IsNSFW,
		arg.PinnedDIDs,
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
			&i.HasVideo,
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
SELECT uri, actor_did, created_at, indexed_at, is_hidden, deleted_at, raw, hashtags, has_media, self_labels, has_video
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
		&i.SelfLabels,
		&i.HasVideo,
	)
	return i, err
}

const listScoredPosts = `-- name: ListScoredPosts :many
SELECT
    cp.uri, cp.actor_did, cp.created_at, cp.indexed_at, cp.is_hidden, cp.deleted_at, cp.raw, cp.hashtags, cp.has_media, cp.self_labels, cp.has_video,
    ph.score
FROM
    candidate_posts AS cp
INNER JOIN candidate_actors AS ca ON cp.actor_did = ca.did
INNER JOIN post_scores AS ph
    ON
        cp.uri = ph.uri AND ph.alg = $1
        AND ph.generation_seq = $2
WHERE
    cp.is_hidden = FALSE
    AND ca.status = 'approved'
    -- Match at least one of the queried hashtags.
    -- If unspecified, do not filter.
    AND (
        COALESCE($3::TEXT [], '{}') = '{}'
        OR $3::TEXT [] && cp.hashtags
    )
    -- If any hashtags are disallowed, filter them out.
    AND (
        COALESCE($4::TEXT [], '{}') = '{}'
        OR NOT $4::TEXT [] && cp.hashtags
    )
    AND (
        -- Match has_media status. If unspecified, do not filter.
        (
            $5::BOOLEAN IS NULL
            OR COALESCE(cp.has_media, FALSE) = $5
        )
        -- Match has_video status. If unspecified, do not filter.
        OR (
            $6::BOOLEAN IS NULL
            OR COALESCE(cp.has_video, FALSE) = $6
        )
    )
    -- Filter by NSFW status. If unspecified, do not filter.
    AND (
        $7::BOOLEAN IS NULL
        OR (
            (ARRAY['nsfw', 'mursuit', 'murrsuit', 'nsfwfurry', 'furrynsfw'] && cp.hashtags)
            OR (ARRAY['porn', 'nudity', 'sexual'] && cp.self_labels)
        ) = $7
    )
    AND cp.deleted_at IS NULL
    AND (
        ROW(ph.score, ph.uri)
        < ROW(($8)::REAL, ($9)::TEXT)
    )
    AND cp.indexed_at > NOW() - INTERVAL '7 day'
ORDER BY
    ph.score DESC, ph.uri DESC
LIMIT $10
`

type ListScoredPostsParams struct {
	Alg                string
	GenerationSeq      int64
	Hashtags           []string
	DisallowedHashtags []string
	HasMedia           pgtype.Bool
	HasVideo           pgtype.Bool
	IsNSFW             pgtype.Bool
	AfterScore         float32
	AfterURI           string
	Limit              int32
}

type ListScoredPostsRow struct {
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
	HasVideo   pgtype.Bool
	Score      float32
}

func (q *Queries) ListScoredPosts(ctx context.Context, arg ListScoredPostsParams) ([]ListScoredPostsRow, error) {
	rows, err := q.db.Query(ctx, listScoredPosts,
		arg.Alg,
		arg.GenerationSeq,
		arg.Hashtags,
		arg.DisallowedHashtags,
		arg.HasMedia,
		arg.HasVideo,
		arg.IsNSFW,
		arg.AfterScore,
		arg.AfterURI,
		arg.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListScoredPostsRow
	for rows.Next() {
		var i ListScoredPostsRow
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
			&i.HasVideo,
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
