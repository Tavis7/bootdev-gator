-- name: CreateFeedFollow :one
WITH feed_followed AS (INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *)
SELECT feed_followed.*, users.name AS username, feeds.name AS feedname
FROM feed_followed
LEFT JOIN users
ON feed_followed.user_id = users.id
LEFT JOIN feeds
ON feed_followed.feed_id = feeds.id;

-- name: GetFeedFollowsForUser :many
SELECT users.name AS username, feeds.name AS feedname, feeds.url AS feed_url
FROM feed_follows
INNER JOIN users
ON feed_follows.user_id = users.id
INNER JOIN feeds
ON feed_follows.feed_id = feeds.id
WHERE users.id = $1;

-- name: DeleteFeedFollow :one
DELETE FROM feed_follows
WHERE feed_follows.user_id = $1 AND feed_follows.feed_id = $2
RETURNING *;
