-- name: GetFurryNewFeed :many
WITH args AS (
    SELECT $7::TEXT [] AS allowed_embeds
)

SELECT cp.uri, cp.actor_did, cp.created_at, cp.indexed_at, cp.is_hidden, cp.deleted_at, cp.raw, cp.hashtags, cp.has_media, cp.self_labels, cp.has_video
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
                COALESCE($1::TEXT [], '{}') = '{}'
                OR $1::TEXT [] && cp.hashtags
            )
            -- If any hashtags are disallowed, filter them out.
            AND (
                COALESCE($2::TEXT [], '{}') = '{}'
                OR NOT $2::TEXT [] && cp.hashtags
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
                $3::BOOLEAN IS NULL
                OR (
                    (ARRAY['nsfw', 'mursuit', 'murrsuit', 'nsfwfurry', 'furrynsfw'] && cp.hashtags)
                    OR (ARRAY['porn', 'nudity', 'sexual'] && cp.self_labels)
                ) = $3
            )
        )
        -- Pinned DID criteria.
        OR cp.actor_did = ANY($4::TEXT [])
    )
    -- Remove posts newer than the cursor timestamp
    AND (cp.indexed_at < $5)
    AND cp.indexed_at > NOW() - INTERVAL '7 day'
    AND cp.created_at > NOW() - INTERVAL '7 day'
ORDER BY
    cp.indexed_at DESC
LIMIT $6
