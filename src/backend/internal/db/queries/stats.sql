-- name: GetWorkspaceStats :one
SELECT
    (SELECT COUNT(*)
    FROM projects p
    WHERE p.workspace_id = $1)::int AS total_projects,
    (SELECT COUNT(*)
    FROM issues i
    JOIN projects p ON i.project_id = p.id
    WHERE p.workspace_id = $1)::int AS total_issues, 
    (SELECT COUNT(*)
    FROM issues i
    JOIN projects p ON i.project_id = p.id
    WHERE p.workspace_id = $1 AND i.assignee_id = $2)::int AS assigned_issues,
    (SELECT COUNT(*)
    FROM issues i
    JOIN projects p ON i.project_id = p.id
    WHERE p.workspace_id = $1 AND i.status = 'done')::int AS completed_issues,
    (SELECT COUNT(*)
    FROM issues i
    JOIN projects p ON i.project_id = p.id
    WHERE p.workspace_id = $1
        AND i.due_date < CURRENT_DATE
        AND i.status != 'done')::int AS overdue_issues;