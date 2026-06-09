package dtos

import "time"

type IssueDto struct {
	ID         string     `json:"id"`
	ProjectID  string     `json:"project_id"`
	Title      string     `json:"title"`
	Overview   string     `json:"overview"`
	Type       string     `json:"type"`
	Status     string     `json:"status"`
	AssigneeID string     `json:"assignee_id"`
	CreatedBy  string     `json:"created_by"`
	DueDate    *time.Time `json:"due_date"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
