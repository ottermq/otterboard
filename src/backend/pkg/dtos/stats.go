package dtos

type StatsDto struct {
	TotalProjects   int32 `json:"total_projects"`
	TotalIssues     int32 `json:"total_issues"`
	AssignedIssues  int32 `json:"assigned_issues"`
	CompletedIssues int32 `json:"completed_issues"`
	OverdueIssues   int32 `json:"overdue_issues"`
}
