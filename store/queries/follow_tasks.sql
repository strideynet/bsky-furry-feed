-- name: EnqueueFollowTask :exec
INSERT INTO follow_tasks (
    actor_did,
    next_try_at,
    created_at,
    should_unfollow
) VALUES ($1, $2, $3, $4);

-- name: GetNextFollowTask :one
SELECT *
FROM follow_tasks
WHERE
    next_try_at <= NOW()
    AND finished_at IS NULL
    AND tries < 3
ORDER BY id ASC
LIMIT 1;

-- name: MarkFollowTaskAsDone :exec
UPDATE follow_tasks
SET finished_at = NOW()
WHERE id = $1;

-- name: MarkFollowTaskAsErrored :exec
UPDATE follow_tasks
SET
    next_try_at = $2,
    tries = tries + 1,
    last_error = $3
WHERE id = $1;
