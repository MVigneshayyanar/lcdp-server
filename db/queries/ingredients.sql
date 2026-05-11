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

-- name: ListIngredientsByMenuItem :many
SELECT i.*, inv.name as inventory_item_name, inv.unit as inventory_item_unit
FROM ingredients i
JOIN inventory_items inv ON i.inventory_item_id = inv.id
WHERE i.menu_item_id = $1;

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

-- name: DeleteIngredientsByMenuItem :exec
DELETE FROM ingredients
WHERE menu_item_id = $1;
