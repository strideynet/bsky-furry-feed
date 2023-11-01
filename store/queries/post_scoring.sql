-- name: DeleteOldPostScores :execrows
DELETE FROM post_scores
WHERE generated_at < sqlc.arg(before)::TIMESTAMPTZ;

-- name: MaterializePostScores :one
WITH seq AS (SELECT NEXTVAL('post_scores_generation_seq') AS seq)

INSERT INTO post_scores (uri, alg, score, generation_seq)
SELECT
    cp.uri AS uri,
    'classic' AS alg,
    (
        SELECT COUNT(*)
        FROM candidate_likes AS cl
        WHERE cl.subject_uri = cp.uri AND cl.deleted_at IS NULL
    )
    / (EXTRACT(EPOCH FROM NOW() - cp.indexed_at) / (60 * 60) + 2)
    ^ 1.85 AS score,
    (SELECT seq FROM seq) AS generation_seq
FROM candidate_posts AS cp
WHERE
    cp.deleted_at IS NULL
    AND cp.indexed_at >= sqlc.arg(after)::TIMESTAMPTZ
RETURNING (SELECT seq FROM seq);

-- name: GetLatestScoreGeneration :one
SELECT ph.generation_seq
FROM post_scores AS ph
WHERE ph.alg = sqlc.arg(alg)
ORDER BY ph.generation_seq DESC
LIMIT 1;
