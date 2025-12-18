-- name: MarkFeedFetched :one
UPDATE feeds
SET last_fetched_at = $1, updated_at = $1
WHERE id = $2
RETURNING *;


-- name: GetStalestFeed :one
SELECT * FROM feeds
ORDER BY last_fetched_at ASC NULLS FIRST LIMIT 1;
