-- name: ListCategories :many
SELECT id, name, slug
FROM categories
ORDER BY name;

-- name: GetCategoriesByPostID :many
SELECT c.id, c.name, c.slug
FROM categories c
INNER JOIN post_categories pc ON pc.category_id = c.id
WHERE pc.post_id = $1
ORDER BY c.name;

-- name: GetOrCreateCategory :one
INSERT INTO categories (name, slug)
VALUES ($1, $2)
ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
RETURNING *;

-- name: ClearPostCategories :exec
DELETE FROM post_categories WHERE post_id = $1;

-- name: AddPostCategory :exec
INSERT INTO post_categories (post_id, category_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;
