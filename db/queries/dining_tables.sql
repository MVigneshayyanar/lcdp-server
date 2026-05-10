-- name: CreateDiningTable :one
INSERT INTO dining_tables (number, status)
VALUES ($1, $2)
RETURNING *;

-- name: GetDiningTable :one
SELECT * FROM dining_tables
WHERE id = $1;

-- name: ListDiningTables :many
SELECT * FROM dining_tables
ORDER BY id;

-- name: UpdateDiningTable :one
UPDATE dining_tables
SET number = $2,
    status = $3,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteDiningTable :exec
DELETE FROM dining_tables
WHERE id = $1;
