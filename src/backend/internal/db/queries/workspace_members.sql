-- name: AddMember :one
INSERT INTO workspace_members (workspace_id, user_id, role)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetMember :one
SELECT * FROM workspace_members
WHERE workspace_id = $1 AND user_id = $2;

-- name: ListMembers :many
SELECT * FROM workspace_members
WHERE workspace_id = $1;

-- name: UpdateMemberRole :one
UPDATE workspace_members SET role = $3
WHERE workspace_id = $1 AND user_id = $2
RETURNING *;

-- name: RemoveMember :exec
DELETE FROM workspace_members
WHERE workspace_id = $1 AND user_id = $2;