-- name: UpdateUserCredentials :one
UPDATE users
SET email = $2,
hashed_password = $3,
updated_at = NOW()
WHERE users.id = $1
RETURNING id, created_at, updated_at, email, is_chirpy_red;