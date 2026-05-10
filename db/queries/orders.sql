-- name: CreateOrder :one
INSERT INTO orders (menu_item_id, quantity, table_id, ordered_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetOrder :one
SELECT * FROM orders
WHERE id = $1;

-- name: ListOrders :many
SELECT * FROM orders
ORDER BY id;

-- name: UpdateOrder :one
UPDATE orders
SET menu_item_id = $2,
    quantity = $3,
    table_id = $4,
    ordered_at = $5,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteOrder :exec
DELETE FROM orders
WHERE id = $1;
