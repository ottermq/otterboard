CREATE TABLE issues (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    overview TEXT, 
    type TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'backlog',
    position FLOAT NOT NULL DEFAULT 0,
    assignee_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    due_date DATE, 
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT issues_type_check CHECK (type IN ('bug', 'task', 'story', 'epic')),
    CONSTRAINT issues_status_check CHECK (status IN ('backlog', 'todo', 'in_progress', 'in_review', 'done'))
);

CREATE INDEX idx_issues_project_id ON issues (project_id);
CREATE INDEX idx_issues_assignee_id ON issues (assignee_id);
CREATE INDEX idx_issues_status ON issues (status);
CREATE INDEX idx_issues_due_date ON issues (due_date);