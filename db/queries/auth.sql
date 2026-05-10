-- name: CreateOTP :one
INSERT INTO otp_codes (user_id, code_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetValidOTP :one
SELECT * FROM otp_codes
WHERE user_id = $1
  AND code_hash = $2
  AND consumed_at IS NULL
  AND expires_at > now()
ORDER BY created_at DESC
LIMIT 1;

-- name: ConsumeOTP :one
UPDATE otp_codes
SET consumed_at = now()
WHERE id = $1
RETURNING *;

-- name: CreateSession :one
INSERT INTO sessions (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetSessionByTokenHash :one
SELECT * FROM sessions
WHERE token_hash = $1
  AND revoked_at IS NULL
  AND expires_at > now();

-- name: UpdateSessionLastSeen :one
UPDATE sessions
SET last_seen_at = now()
WHERE id = $1
RETURNING *;

-- name: RevokeSession :one
UPDATE sessions
SET revoked_at = now()
WHERE id = $1
RETURNING *;
