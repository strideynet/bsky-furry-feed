-- name: CreateCandidatePost :exec
INSERT INTO
candidate_posts (
    uri, actor_did, created_at, indexed_at, hashtags, has_media, raw, self_labels
)
VALUES
($1, $2, $3, $4, $5, $6, $7, $8);

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
      -- Only include posts by approved actors
      ca.status = 'approved'
      -- Remove posts hidden by our moderators
  AND cp.is_hidden = false
      -- Remove posts deleted by the actors
  AND cp.deleted_at IS NULL
      -- Match at least one of the queried hashtags. If unspecified, do not filter.
  AND (
    -- Standard criteria.
    (
        (COALESCE(@hashtags::TEXT[], '{}') = '{}' OR
            @hashtags::TEXT[] && cp.hashtags)
            -- Match has_media status. If unspecified, do not filter.
        AND (sqlc.narg(has_media)::BOOLEAN IS NULL OR
            COALESCE(cp.has_media, false) = @has_media)
            -- Filter by NSFW status. If unspecified, do not filter.
        AND (sqlc.narg(is_nsfw)::BOOLEAN IS NULL OR
            ((ARRAY ['nsfw', 'mursuit', 'murrsuit'] && cp.hashtags) OR
                (ARRAY ['porn', 'nudity', 'sexual'] && cp.self_labels)) = @is_nsfw)
    ) OR
    -- Pinned DID criteria.
    cp.actor_did = ANY(@pinned_dids::TEXT[])
  )
      -- Remove posts newer than the cursor timestamp
  AND (cp.indexed_at < @cursor_timestamp)
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

-- name: ListScoredPosts :many
SELECT
    cp.*,
    ph.score
FROM
    candidate_posts AS cp
INNER JOIN candidate_actors AS ca ON cp.actor_did = ca.did
INNER JOIN post_scores AS ph
    ON
        ph.uri = cp.uri AND ph.alg = sqlc.arg(alg)
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
