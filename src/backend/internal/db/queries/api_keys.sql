-- name: CreateApiKey :one
INSERT INTO api_keys (workspace_id, user_id, name, key_hash)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListApiKeys :many
SELECT * FROM api_keys
WHERE workspace_id = $1 AND revoked_at IS NULL;

-- name: GetApiKeyByHash :one
SELECT * FROM api_keys
WHERE key_hash = $1 AND revoked_at IS NULL;

-- name: RevokeApiKey :one
UPDATE api_keys SET revoked_at = NOW()
WHERE id = $1 AND workspace_id = $2
RETURNING *;