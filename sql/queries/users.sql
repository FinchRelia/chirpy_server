-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: DeleteUser :one
DELETE FROM users
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE users.email = $1;

-- name: UpgradeUser :exec
UPDATE users
SET is_chirpy_red = true
WHERE id = $1;

-- name: UpdateUserCredentials :one
UPDATE users
SET email = $2,
hashed_password = $3,
updated_at = NOW()
WHERE users.id = $1
RETURNING id, created_at, updated_at, email, is_chirpy_red;