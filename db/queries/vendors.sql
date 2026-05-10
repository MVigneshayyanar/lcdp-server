-- name: CreateVendor :one
INSERT INTO vendors (name, address, email, phone)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetVendor :one
SELECT * FROM vendors
WHERE id = $1;

-- name: ListVendors :many
SELECT * FROM vendors
ORDER BY id;

-- name: UpdateVendor :one
UPDATE vendors
SET name = $2,
    address = $3,
    email = $4,
    phone = $5,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteVendor :exec
DELETE FROM vendors
WHERE id = $1;
