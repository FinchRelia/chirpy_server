-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1
)
RETURNING *;

-- name: DeleteUser :one
DELETE FROM users
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE users.email = $1;