-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
    id,
    user_id,
    token_hash,
    family_id,
    expires_at,
    revoked_at,
    created_at
)
VALUES (
    sqlc.arg(id),
    sqlc.arg(user_id),
    sqlc.arg(token_hash),
    sqlc.arg(family_id),
    sqlc.arg(expires_at),
    NULL,
    sqlc.arg(created_at)
)
RETURNING *;

-- name: GetRefreshTokenByHash :one
SELECT *
FROM refresh_tokens
WHERE token_hash = sqlc.arg(token_hash);

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = sqlc.arg(revoked_at)
WHERE id = sqlc.arg(id);

-- name: ConsumeRefreshToken :one
UPDATE refresh_tokens
SET revoked_at = sqlc.arg(revoked_at)
WHERE id = sqlc.arg(id) AND revoked_at IS NULL
RETURNING *;

-- name: RevokeTokenFamily :exec
UPDATE refresh_tokens
SET revoked_at = sqlc.arg(revoked_at)
WHERE family_id = sqlc.arg(family_id) AND revoked_at IS NULL;
