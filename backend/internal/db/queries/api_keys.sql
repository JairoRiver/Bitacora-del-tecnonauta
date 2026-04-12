-- name: GetAPIKey :one
SELECT * FROM api_keys LIMIT 1;

-- name: CreateAPIKey :one
INSERT INTO api_keys (key_hash)
VALUES ($1)
RETURNING *;

-- name: DeleteAPIKey :exec
DELETE FROM api_keys;
