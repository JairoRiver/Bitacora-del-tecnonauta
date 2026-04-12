-- name: GetContentBlocksByPostID :many
SELECT id, post_id, position, type, value, conf
FROM content_blocks
WHERE post_id = $1
ORDER BY position;

-- name: CreateContentBlock :one
INSERT INTO content_blocks (post_id, position, type, value, conf)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateContentBlock :one
UPDATE content_blocks
SET position = $2, type = $3, value = $4, conf = $5
WHERE id = $1
RETURNING *;

-- name: DeleteContentBlock :exec
DELETE FROM content_blocks WHERE id = $1;

-- name: DeleteContentBlocksByPostID :exec
DELETE FROM content_blocks WHERE post_id = $1;
