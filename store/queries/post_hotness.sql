-- name: DeleteOldPostHotness :execrows
DELETE FROM post_hotness
WHERE generated_at < NOW() - sqlc.arg(retention_period)::INTERVAL;

-- name: MaterializeClassicPostHotness :one
WITH seq AS (SELECT NEXTVAL('post_hotness_generation_seq') AS seq)

INSERT INTO post_hotness (uri, alg, score, generation_seq)
SELECT
    cp.uri AS uri,
    'classic' AS alg,
    (
        SELECT COUNT(*)
        FROM candidate_likes AS cl
        WHERE cl.subject_uri = cp.uri AND cl.deleted_at IS NULL
    )
    / (EXTRACT(EPOCH FROM NOW() - cp.created_at) / (60 * 60) + 2)
    ^ 1.85 AS score,
    (SELECT seq FROM seq) AS generation_seq
FROM candidate_posts AS cp
WHERE
    cp.deleted_at IS NULL
    AND cp.created_at >= NOW() - sqlc.arg(lookback_period)::INTERVAL
RETURNING (SELECT seq FROM seq);
