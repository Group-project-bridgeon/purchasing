-- name: CreateOutboxEvent :one
INSERT INTO outbox_events (
    aggregate_type,
    aggregate_id,
    event_type,
    payload
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetPendingOutboxEvents :many
SELECT * FROM outbox_events
WHERE status = 'pending'
ORDER BY created_at ASC
LIMIT $1;

-- name: MarkOutboxEventPublished :exec
UPDATE outbox_events
SET status = 'published',
    published_at = NOW(),
    last_error = NULL
WHERE id = $1;

-- name: MarkOutboxEventFailed :exec
UPDATE outbox_events
SET retry_count = retry_count + 1,
    last_error = $2,
    status = CASE WHEN retry_count + 1 >= 5 THEN 'failed'::varchar(20) ELSE 'pending'::varchar(20) END
WHERE id = $1;

-- name: ListOutboxEvents :many
SELECT * FROM outbox_events
WHERE (sqlc.narg('status')::varchar IS NULL OR status = sqlc.narg('status'))
  AND (sqlc.narg('aggregate_type')::varchar IS NULL OR aggregate_type = sqlc.narg('aggregate_type'))
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetOutboxEventByID :one
SELECT * FROM outbox_events
WHERE id = $1;

-- name: RetryOutboxEvent :exec
UPDATE outbox_events
SET status = 'pending',
    retry_count = 0,
    last_error = NULL
WHERE id = $1;
