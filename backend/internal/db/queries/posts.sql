-- name: ListPosts :many
SELECT id, title, subtitle, date, hero_image, slug, created_at, updated_at
FROM posts
ORDER BY date DESC;

-- name: GetPostBySlug :one
SELECT id, title, subtitle, date, hero_image, slug, created_at, updated_at
FROM posts
WHERE slug = $1;

-- name: GetPostByID :one
SELECT id, title, subtitle, date, hero_image, slug, created_at, updated_at
FROM posts
WHERE id = $1;

-- name: ListPostsByCategory :many
SELECT DISTINCT p.id, p.title, p.subtitle, p.date, p.hero_image, p.slug, p.created_at, p.updated_at
FROM posts p
INNER JOIN post_categories pc ON pc.post_id = p.id
INNER JOIN categories c ON c.id = pc.category_id
WHERE c.slug = $1
ORDER BY p.date DESC;

-- name: CreatePost :one
INSERT INTO posts (title, subtitle, date, hero_image, slug)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdatePost :one
UPDATE posts
SET title = $2, subtitle = $3, date = $4, hero_image = $5, slug = $6, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeletePost :exec
DELETE FROM posts WHERE id = $1;
