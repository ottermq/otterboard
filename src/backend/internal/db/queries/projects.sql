-- name: CreateProject :one
INSERT INTO projects (workspace_id, name, image_url)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetProjectByID :one
SELECT * FROM projects
WHERE id = $1 AND workspace_id = $2;

-- name: ListProjectsByWorkspace :many
SELECT * FROM projects
WHERE workspace_id = $1
ORDER BY created_at DESC 
LIMIT $2 OFFSET $3;

-- name: CountProjectsByWorkspace :one
SELECT COUNT(*) FROM projects
WHERE workspace_id = $1;

-- name: UpdateProject :one
UPDATE projects SET 
name = $3, 
image_url = $4,
updated_at = NOW()
WHERE id = $1 AND workspace_id = $2
RETURNING *;

-- name: DeleteProject :exec
DELETE FROM projects
WHERE id = $1 AND workspace_id = $2;