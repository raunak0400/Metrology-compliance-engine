-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, jti, token_hash, expires_at, user_agent, ip)
VALUES ($1, $2, $3, $4, sqlc.narg('user_agent'), sqlc.narg('ip'))
RETURNING *;

-- name: GetRefreshTokenByJTI :one
SELECT * FROM refresh_tokens WHERE jti = $1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens SET revoked_at = now() WHERE jti = $1;

-- name: RevokeRefreshTokenIfActive :exec
UPDATE refresh_tokens SET revoked_at = now() WHERE jti = $1 AND revoked_at IS NULL;
