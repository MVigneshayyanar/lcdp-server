-- name: CreateMenuItem :one
INSERT INTO menu_items (name, price, is_available, category, description)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetMenuItem :one
SELECT * FROM menu_items
WHERE id = $1;

-- name: ListMenuItems :many
SELECT * FROM menu_items
ORDER BY id;

-- name: UpdateMenuItem :one
UPDATE menu_items
SET name = $2,
    price = $3,
    is_available = $4,
    category = $5,
    description = $6,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteMenuItem :exec
DELETE FROM menu_items
WHERE id = $1;
