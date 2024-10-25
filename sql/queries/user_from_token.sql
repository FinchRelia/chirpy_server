-- name: GetUserFromRefreshToken :one
SELECT user_id FROM refresh_tokens
WHERE refresh_tokens.token = $1
AND refresh_tokens.expires_at > NOW()
AND refresh_tokens.revoked_at IS NULL;