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

-- name: ListIssuesByProjectFiltered :many
SELECT * FROM issues
WHERE project_id = @project_id
    AND (sqlc.narg(status)::text  IS NULL OR status   = sqlc.narg(status))
    AND (sqlc.narg(type)::text     IS NULL OR type     = sqlc.narg(type))
    AND (@assignee_id::uuid     IS NULL OR assignee_id  = @assignee_id)
    AND (@due_before::date      IS NULL OR due_date <= @due_before)
    AND (@due_after::date         IS NULL OR due_date >= @due_after)
ORDER BY
    CASE WHEN @sort_by::text = 'title'              AND @sort_order::text = 'asc'       THEN title              END ASC       NULLS LAST,
    CASE WHEN @sort_by::text = 'title'              AND @sort_order::text = 'desc'     THEN title              END DESC     NULLS LAST,
    CASE WHEN @sort_by::text = 'status'          AND @sort_order::text = 'asc'       THEN status          END ASC       NULLS LAST,
    CASE WHEN @sort_by::text = 'status'          AND @sort_order::text = 'desc'     THEN status          END DESC     NULLS LAST,
    CASE WHEN @sort_by::text = 'due_date'     AND @sort_order::text = 'asc'       THEN due_date     END ASC       NULLS LAST,
    CASE WHEN @sort_by::text = 'due_date'     AND @sort_order::text = 'desc'     THEN due_date     END DESC     NULLS LAST,
    CASE WHEN @sort_by::text = 'created_at'   AND @sort_order::text = 'asc'       THEN created_at   END ASC       NULLS LAST,
    CASE WHEN @sort_by::text = 'created_at'   AND @sort_order::text = 'desc'     THEN created_at   END DESC     NULLS LAST,
    position ASC
LIMIT @_limit OFFSET @_offset;

-- name: CountIssuesByProjectFiltered :one
SELECT COUNT(*) FROM issues
WHERE project_id = @project_id
    AND (sqlc.narg(status)::text  IS NULL OR status   = sqlc.narg(status))
    AND (sqlc.narg(type)::text     IS NULL OR type     = sqlc.narg(type))
    AND (@assignee_id::uuid     IS NULL OR assignee_id  = @assignee_id)
    AND (@due_before::date      IS NULL OR due_date <= @due_before)
    AND (@due_after::date         IS NULL OR due_date >= @due_after);

-- name: ListIssuesByWorkspaceFiltered :many
SELECT issues.* FROM issues
JOIN projects ON issues.project_id = projects.id
WHERE projects.workspace_id = @workspace_id
    AND (@project_id_filter::uuid IS NULL OR issues.project_id = @project_id_filter)
    AND (sqlc.narg(status)::text  IS NULL OR issues.status   = sqlc.narg(status))
    AND (sqlc.narg(type)::text     IS NULL OR issues.type     = sqlc.narg(type))
    AND (@assignee_id::uuid     IS NULL OR issues.assignee_id  = @assignee_id)
    AND (@due_before::date      IS NULL OR issues.due_date <= @due_before)
    AND (@due_after::date         IS NULL OR issues.due_date >= @due_after)
ORDER BY
    CASE WHEN @sort_by::text = 'title'              AND @sort_order::text = 'asc'       THEN issues.title              END ASC       NULLS LAST,
    CASE WHEN @sort_by::text = 'title'              AND @sort_order::text = 'desc'     THEN issues.title              END DESC     NULLS LAST,
    CASE WHEN @sort_by::text = 'status'          AND @sort_order::text = 'asc'       THEN issues.status          END ASC       NULLS LAST,
    CASE WHEN @sort_by::text = 'status'          AND @sort_order::text = 'desc'     THEN issues.status          END DESC     NULLS LAST,
    CASE WHEN @sort_by::text = 'due_date'     AND @sort_order::text = 'asc'       THEN issues.due_date     END ASC       NULLS LAST,
    CASE WHEN @sort_by::text = 'due_date'     AND @sort_order::text = 'desc'     THEN issues.due_date     END DESC     NULLS LAST,
    CASE WHEN @sort_by::text = 'created_at'   AND @sort_order::text = 'asc'       THEN issues.created_at   END ASC       NULLS LAST,
    CASE WHEN @sort_by::text = 'created_at'   AND @sort_order::text = 'desc'     THEN issues.created_at   END DESC     NULLS LAST,
    position ASC
LIMIT @_limit OFFSET @_offset;

-- name: CountIssuesByWorkspaceFiltered :one
SELECT COUNT(*) FROM issues
JOIN projects ON issues.project_id = projects.id
WHERE projects.workspace_id = @workspace_id
    AND (@project_id_filter::uuid IS NULL OR issues.project_id = @project_id_filter)
    AND (sqlc.narg(status)::text  IS NULL OR issues.status   = sqlc.narg(status))
    AND (sqlc.narg(type)::text     IS NULL OR issues.type     = sqlc.narg(type))
    AND (@assignee_id::uuid     IS NULL OR issues.assignee_id  = @assignee_id)
    AND (@due_before::date      IS NULL OR issues.due_date <= @due_before)
    AND (@due_after::date         IS NULL OR issues.due_date >= @due_after);

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