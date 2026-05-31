-- name: GetMaxPositionByProjectAndStatus :one
SELECT COALESCE(MAX(position), 0) FROM issues 
WHERE project_id = $1 AND status = $2;

-- name: CreateIssue :one
INSERT INTO issues (project_id, title, overview, type, status, position, assignee_id, created_by, due_date)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetIssueByID :one
SELECT * FROM issues
WHERE id = $1 AND project_id = $2;

-- name: ListIssuesByProject :many 
SELECT * FROM issues
WHERE project_id = $1
ORDER BY position ASC 
LIMIT $2 OFFSET $3;

-- name: CountIssuesByProject :one
SELECT COUNT(*) FROM issues
WHERE project_id = $1;

-- name: ListIssuesByWorkspace :many
SELECT issues.* FROM issues
JOIN projects ON issues.project_id = projects.id
WHERE projects.workspace_id = $1
ORDER BY issues.position ASC 
LIMIT $2 OFFSET $3;

-- name: CountIssuesByWorkspace :one
SELECT COUNT(*) FROM issues
JOIN projects ON issues.project_id = projects.id
WHERE projects.workspace_id = $1;

-- name: UpdateIssue :one  
UPDATE issues SET
title = $3, 
overview = $4,
type = $5,
status = $6,
position = $7,
assignee_id = $8,
due_date = $9,
updated_at = NOW() 
WHERE id = $1 AND project_id = $2
RETURNING *;

-- name: DeleteIssue :exec       
DELETE FROM issues
WHERE id = $1 AND project_id = $2;