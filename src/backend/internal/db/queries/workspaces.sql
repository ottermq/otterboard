-- name: CreateWorkspace :one
INSERT INTO workspaces (name, owner_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetWorkspaceByID :one
SELECT * FROM workspaces
WHERE id = $1;

-- name: GetWorkspacesByOwnerID :many
SELECT * FROM workspaces
WHERE owner_id = $1;

-- name: UpdateWorkspace :one
UPDATE workspaces SET name = $1, 
updated_at = NOW()
WHERE id = $2
RETURNING *;

-- name: DeleteWorkspace :exec
DELETE FROM workspaces
WHERE id = $1;

-- name: GetWorkspaceByMemberID :many
SELECT w.* FROM workspaces w
JOIN workspace_members wm ON w.id = wm.workspace_id
WHERE wm.user_id = $1;