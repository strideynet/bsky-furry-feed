-- name: CreateCandidatePost :exec
INSERT INTO
candidate_posts (
    uri, actor_did, created_at, indexed_at, hashtags, has_media, raw
)
VALUES
($1, $2, $3, $4, $5, $6, $7);

-- name: SoftDeleteCandidatePost :exec
UPDATE
candidate_posts
SET
    deleted_at = NOW()
WHERE
    uri = $1;

-- name: GetFurryNewFeed :many
SELECT cp.*
FROM
    candidate_posts AS cp
INNER JOIN candidate_actors AS ca ON cp.actor_did = ca.did
WHERE
    cp.is_hidden = false
    AND ca.status = 'approved'
    AND (
        COALESCE(sqlc.arg(hashtags)::TEXT [], '{}') = '{}'
        OR sqlc.arg(hashtags)::TEXT [] && cp.hashtags
    )
    AND (
        sqlc.narg(has_media)::BOOLEAN IS NULL
        OR COALESCE(cp.has_media, false) = sqlc.arg(has_media)
    )
    AND (
        sqlc.narg(is_nsfw)::BOOLEAN IS NULL
        OR (ARRAY['nsfw', 'mursuit', 'murrsuit'] && cp.hashtags)
        = sqlc.arg(is_nsfw)
    )
    AND (cp.indexed_at < sqlc.arg(cursor_timestamp))
    AND cp.deleted_at IS NULL
ORDER BY
    cp.indexed_at DESC
LIMIT sqlc.arg(_limit);

-- name: GetPostsWithLikes :many
SELECT
    cp.*,
    (
        SELECT COUNT(*)
        FROM
            candidate_likes AS cl
        WHERE
            cl.subject_uri = cp.uri
            AND (cl.indexed_at < sqlc.arg(cursor_timestamp))
            AND cl.deleted_at IS NULL
    ) AS likes
FROM
    candidate_posts AS cp
INNER JOIN candidate_actors AS ca ON cp.actor_did = ca.did
WHERE
    cp.is_hidden = false
    AND ca.status = 'approved'
    AND (
        sqlc.arg(cursor_timestamp)::TIMESTAMPTZ IS NULL
        OR cp.indexed_at < sqlc.arg(cursor_timestamp)
    )
    AND cp.deleted_at IS NULL
ORDER BY
    cp.indexed_at DESC
LIMIT sqlc.arg(_limit);

-- name: GetPostByURI :one
SELECT *
FROM
    candidate_posts AS cp
WHERE
    cp.uri = sqlc.arg(uri)
LIMIT 1;

-- name: GetLatestHotPostGeneration :one
SELECT ph.generation_seq
FROM post_hotness AS ph
WHERE ph.alg = sqlc.arg(alg)
ORDER BY ph.generation_seq DESC
LIMIT 1;

-- name: GetHotPosts :many
SELECT
    cp.*,
    ph.score
FROM
    candidate_posts AS cp
INNER JOIN candidate_actors AS ca ON cp.actor_did = ca.did
INNER JOIN post_hotness AS ph
    ON
        ph.post_uri = cp.uri AND ph.alg = sqlc.arg(alg)
        AND ph.generation_seq = sqlc.arg(generation_seq)
WHERE
    cp.is_hidden = false
    AND ca.status = 'approved'
    AND (
        COALESCE(sqlc.arg(hashtags)::TEXT [], '{}') = '{}'
        OR sqlc.arg(hashtags)::TEXT [] && cp.hashtags
    )
    AND (
        sqlc.narg(has_media)::BOOLEAN IS NULL
        OR COALESCE(cp.has_media, false) = sqlc.narg(has_media)
    )
    AND (
        sqlc.narg(is_nsfw)::BOOLEAN IS NULL
        OR (ARRAY['nsfw', 'mursuit', 'murrsuit'] && cp.hashtags)
        = sqlc.narg(is_nsfw)
    )
    AND cp.deleted_at IS NULL
    AND (
        (ph.score, ph.uri)
        < (sqlc.arg(after_score)::REAL, sqlc.arg(after_uri)::TEXT)
    )
ORDER BY
    ph.score DESC, ph.uri DESC
LIMIT sqlc.arg(_limit);
