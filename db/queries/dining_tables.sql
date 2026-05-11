-- name: CreateDiningTable :one
INSERT INTO dining_tables (number)
VALUES ($1)
RETURNING *;

-- name: GetDiningTable :one
SELECT * FROM dining_tables
WHERE id = $1;

-- name: ListDiningTables :many
SELECT * FROM dining_tables
ORDER BY number;

-- name: DeleteDiningTable :exec
DELETE FROM dining_tables
WHERE id = $1;
