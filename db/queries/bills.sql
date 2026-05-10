-- name: CreateBill :one
INSERT INTO bills (vendor_id, txn_id, amount, due_date, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetBill :one
SELECT * FROM bills
WHERE id = $1;

-- name: ListBills :many
SELECT * FROM bills
ORDER BY id;

-- name: UpdateBill :one
UPDATE bills
SET vendor_id = $2,
    txn_id = $3,
    amount = $4,
    due_date = $5,
    status = $6,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteBill :exec
DELETE FROM bills
WHERE id = $1;
