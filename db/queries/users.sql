-- name: CreateUser :one
INSERT INTO users (name, role, phone)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByPhone :one
SELECT * FROM users
WHERE phone = $1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY id;

-- name: UpdateUser :one
UPDATE users
SET name = $2,
    role = $3,
    phone = $4,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
