-- name: CreateIngredient :one
INSERT INTO ingredients (menu_item_id, inventory_item_id, quantity)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetIngredient :one
SELECT * FROM ingredients
WHERE id = $1;

-- name: ListIngredients :many
SELECT * FROM ingredients
ORDER BY id;

-- name: UpdateIngredient :one
UPDATE ingredients
SET menu_item_id = $2,
    inventory_item_id = $3,
    quantity = $4,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteIngredient :exec
DELETE FROM ingredients
WHERE id = $1;
