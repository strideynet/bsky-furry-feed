-- name: CreateCandidatePost :exec
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
($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: SoftDeleteCandidatePost :exec
UPDATE
candidate_posts
SET
    deleted_at = NOW()
WHERE
    uri = $1;

-- name: GetFurryNewFeed :many
WITH args AS (
    SELECT sqlc.narg(allowed_embeds)::TEXT [] AS allowed_embeds
)

SELECT cp.*
FROM
    candidate_posts AS cp
INNER JOIN candidate_actors AS ca ON cp.actor_did = ca.did
NATURAL JOIN args
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
                COALESCE(sqlc.narg(hashtags)::TEXT [], '{}') = '{}'
                OR sqlc.arg(hashtags)::TEXT [] && cp.hashtags
            )
            -- If any hashtags are disallowed, filter them out.
            AND (
                COALESCE(sqlc.narg(disallowed_hashtags)::TEXT [], '{}') = '{}'
                OR NOT sqlc.narg(disallowed_hashtags)::TEXT [] && cp.hashtags
            )
            AND (
                CARDINALITY(args.allowed_embeds) = 0
                OR (
                    'none' = ANY(args.allowed_embeds)
                    AND COALESCE(cp.has_media, FALSE) = FALSE
                    AND COALESCE(cp.has_video, FALSE) = FALSE
                )
                OR (
                    'image' = ANY(args.allowed_embeds)
                    AND COALESCE(cp.has_media, FALSE) = TRUE
                )
                OR (
                    'video' = ANY(args.allowed_embeds)
                    AND COALESCE(cp.has_video, FALSE) = TRUE
                )
            )
            -- Filter by NSFW status. If unspecified, do not filter.
            AND (
                sqlc.narg(is_nsfw)::BOOLEAN IS NULL
                OR (
                    (ARRAY['nsfw', 'mursuit', 'murrsuit', 'nsfwfurry', 'furrynsfw'] && cp.hashtags)
                    OR (ARRAY['porn', 'nudity', 'sexual'] && cp.self_labels)
                ) = sqlc.narg(is_nsfw)
            )
        )
        -- Pinned DID criteria.
        OR cp.actor_did = ANY(sqlc.arg(pinned_dids)::TEXT [])
    )
    -- Remove posts newer than the cursor timestamp
    AND (cp.indexed_at < sqlc.arg(cursor_timestamp))
    AND cp.indexed_at > NOW() - INTERVAL '7 day'
    AND cp.created_at > NOW() - INTERVAL '7 day'
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
WITH args AS (
    SELECT sqlc.narg(allowed_embeds)::TEXT [] AS allowed_embeds
)

SELECT
    cp.*,
    ph.score
FROM
    candidate_posts AS cp
INNER JOIN candidate_actors AS ca ON cp.actor_did = ca.did
INNER JOIN post_scores AS ph
    ON
        cp.uri = ph.uri AND ph.alg = sqlc.arg(alg)
        AND ph.generation_seq = sqlc.arg(generation_seq)
NATURAL JOIN args
WHERE
    cp.is_hidden = FALSE
    AND ca.status = 'approved'
    -- Match at least one of the queried hashtags.
    -- If unspecified, do not filter.
    AND (
        COALESCE(sqlc.narg(hashtags)::TEXT [], '{}') = '{}'
        OR sqlc.narg(hashtags)::TEXT [] && cp.hashtags
    )
    -- If any hashtags are disallowed, filter them out.
    AND (
        COALESCE(sqlc.narg(disallowed_hashtags)::TEXT [], '{}') = '{}'
        OR NOT sqlc.narg(disallowed_hashtags)::TEXT [] && cp.hashtags
    )
    AND (
        CARDINALITY(args.allowed_embeds) = 0
        OR (
            'none' = ANY(args.allowed_embeds)
            AND COALESCE(cp.has_media, FALSE) = FALSE
            AND COALESCE(cp.has_video, FALSE) = FALSE
        )
        OR (
            'image' = ANY(args.allowed_embeds)
            AND COALESCE(cp.has_media, FALSE) = TRUE
        )
        OR (
            'video' = ANY(args.allowed_embeds)
            AND COALESCE(cp.has_video, FALSE) = TRUE
        )
    )
    -- Filter by NSFW status. If unspecified, do not filter.
    AND (
        sqlc.narg(is_nsfw)::BOOLEAN IS NULL
        OR (
            (ARRAY['nsfw', 'mursuit', 'murrsuit', 'nsfwfurry', 'furrynsfw'] && cp.hashtags)
            OR (ARRAY['porn', 'nudity', 'sexual'] && cp.self_labels)
        ) = sqlc.narg(is_nsfw)
    )
    AND cp.deleted_at IS NULL
    AND (
        ROW(ph.score, ph.uri)
        < ROW((sqlc.arg(after_score))::REAL, (sqlc.arg(after_uri))::TEXT)
    )
    AND cp.indexed_at > NOW() - INTERVAL '7 day'
    AND cp.created_at > NOW() - INTERVAL '7 day'
ORDER BY
    ph.score DESC, ph.uri DESC
LIMIT sqlc.arg(_limit);
