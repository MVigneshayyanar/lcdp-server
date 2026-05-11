-- name: CreateInventoryItem :one
INSERT INTO inventory_items (name, quantity, unit, vendor_id, category, min_stock)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetInventoryItem :one
SELECT * FROM inventory_items
WHERE id = $1;

-- name: ListInventoryItems :many
SELECT * FROM inventory_items
ORDER BY id;

-- name: UpdateInventoryItem :one
UPDATE inventory_items
SET name = $2,
    quantity = $3,
    unit = $4,
    vendor_id = $5,
    category = $6,
    min_stock = $7,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteInventoryItem :exec
DELETE FROM inventory_items
WHERE id = $1;
