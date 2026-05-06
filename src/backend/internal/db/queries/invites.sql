-- name: CreateInvite :one
INSERT INTO invites (workspace_id, created_by, token, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetInviteByToken :one
SELECT * FROM invites
WHERE token = $1;

-- name: UseInvite :one
UPDATE invites SET used_at = NOW()
WHERE token = $1 AND used_at IS NULL
RETURNING *;