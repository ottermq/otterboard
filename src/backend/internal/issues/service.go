package issues

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ottermq/otterboard/src/backend/internal/common"
	"github.com/ottermq/otterboard/src/backend/internal/db"
)

var (
	ErrInvalidTitle      = common.NewAppError(http.StatusBadRequest, "invalid issue title")
	ErrInvalidType       = common.NewAppError(http.StatusBadRequest, "invalid issue type")
	ErrInvalidStatus     = common.NewAppError(http.StatusBadRequest, "invalid issue status")
	ErrInvalidAssigneeID = common.NewAppError(http.StatusBadRequest, "invalid assignee ID")
)

const (
	DefaultIssueStatus = "backlog"
)

type Issue struct {
	ID         string
	ProjectID  string
	Title      string
	Overview   string
	Type       string
	Status     string
	Position   float64
	AssigneeID string
	CreatedBy  string
	DueDate    *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

var validTypes = map[string]struct{}{
	"bug":   {},
	"task":  {},
	"story": {},
	"epic":  {},
}

func IsValidType(t string) bool {
	_, ok := validTypes[t]
	return ok
}

var validStatuses = map[string]struct{}{
	"backlog":     {},
	"todo":        {},
	"in_progress": {},
	"in_review":   {},
	"done":        {},
}

func IsValidStatus(s string) bool {
	_, ok := validStatuses[s]
	return ok
}

type IssueStore interface {
	GetMaxPositionByProjectAndStatus(ctx context.Context, arg db.GetMaxPositionByProjectAndStatusParams) (any, error)
	CreateIssue(ctx context.Context, arg db.CreateIssueParams) (db.Issue, error)
	GetIssueByID(ctx context.Context, arg db.GetIssueByIDParams) (db.Issue, error)
	ListIssuesByProject(ctx context.Context, arg db.ListIssuesByProjectParams) ([]db.Issue, error)
	CountIssuesByProject(ctx context.Context, projectID pgtype.UUID) (int64, error)
	ListIssuesByWorkspace(ctx context.Context, arg db.ListIssuesByWorkspaceParams) ([]db.Issue, error)
	CountIssuesByWorkspace(ctx context.Context, workspaceID pgtype.UUID) (int64, error)
	UpdateIssue(ctx context.Context, arg db.UpdateIssueParams) (db.Issue, error)
	DeleteIssue(ctx context.Context, arg db.DeleteIssueParams) error
}

type CreateIssueInput struct {
	ProjectID  string
	Title      string
	Overview   string
	Type       string
	AssigneeID string
	CreatedBy  string
	DueDate    *time.Time
}

type GetMaxPositionByProjectAndStatusInput struct {
	ProjectID string
	Status    string
}

type IssueService struct {
	store IssueStore
}

func NewIssueService(store IssueStore) *IssueService {
	return &IssueService{
		store: store,
	}
}

func (i *IssueService) CreateIssue(ctx context.Context, input CreateIssueInput) (Issue, error) {
	var projectUUID pgtype.UUID
	if err := projectUUID.Scan(input.ProjectID); err != nil {
		return Issue{}, common.ErrInvalidProjectID
	}
	var assigneeUUID pgtype.UUID
	if input.AssigneeID != "" {
		if err := assigneeUUID.Scan(input.AssigneeID); err != nil {
			return Issue{}, ErrInvalidAssigneeID
		}
	}
	var createdByUUID pgtype.UUID
	if input.CreatedBy != "" {
		if err := createdByUUID.Scan(input.CreatedBy); err != nil {
			return Issue{}, common.ErrInvalidUserID
		}
	}
	if input.Title == "" {
		return Issue{}, ErrInvalidTitle
	}
	var overview pgtype.Text
	if input.Overview == "" {
		overview = pgtype.Text{Valid: false}
	} else {
		overview = pgtype.Text{
			String: input.Overview,
			Valid:  true}
	}
	if !IsValidType(input.Type) {
		return Issue{}, ErrInvalidType
	}
	status := DefaultIssueStatus
	maxPos, err := i.getMaxPositionByProjectAndStatus(ctx, projectUUID, status)
	if err != nil {
		return Issue{}, err
	}
	position := maxPos + 1000.0
	issue, err := i.store.CreateIssue(ctx, db.CreateIssueParams{
		ProjectID:  projectUUID,
		Title:      input.Title,
		Overview:   overview,
		Type:       input.Type,
		Status:     status,
		Position:   position,
		AssigneeID: assigneeUUID,
		CreatedBy:  createdByUUID,
		DueDate:    dueDateToPgDate(input.DueDate),
	})
	if err != nil {
		return Issue{}, err
	}
	return mapToIssueDomain(issue), nil
}

func (i *IssueService) GetMaxPositionByProjectAndStatus(ctx context.Context, input GetMaxPositionByProjectAndStatusInput) (float64, error) {
	var projectUUID pgtype.UUID
	if err := projectUUID.Scan(input.ProjectID); err != nil {
		return 0.0, common.ErrInvalidProjectID
	}
	if !IsValidStatus(input.Status) {
		return 0.0, ErrInvalidStatus
	}
	return i.getMaxPositionByProjectAndStatus(ctx, projectUUID, input.Status)
}

func (i *IssueService) getMaxPositionByProjectAndStatus(ctx context.Context, projectUUID pgtype.UUID, status string) (float64, error) {
	raw, err := i.store.GetMaxPositionByProjectAndStatus(ctx, db.GetMaxPositionByProjectAndStatusParams{
		ProjectID: projectUUID,
		Status:    status,
	})
	if err != nil {
		return 0.0, err
	}
	maxPos, _ := raw.(float64)
	return maxPos, nil
}

func dueDateToPgDate(dueDate *time.Time) pgtype.Date {
	if dueDate == nil {
		return pgtype.Date{Valid: false}
	}
	return pgtype.Date{
		Time:  *dueDate,
		Valid: true,
	}
}

func mapToIssueDomain(issue db.Issue) Issue {
	overview := ""
	if issue.Overview.Valid {
		overview = issue.Overview.String
	}
	var dueDate *time.Time
	if issue.DueDate.Valid {
		t := issue.DueDate.Time
		dueDate = &t
	}

	return Issue{
		ID:         issue.ID.String(),
		ProjectID:  issue.ProjectID.String(),
		Title:      issue.Title,
		Overview:   overview,
		Type:       issue.Type,
		Status:     issue.Status,
		Position:   issue.Position,
		AssigneeID: issue.AssigneeID.String(),
		CreatedBy:  issue.CreatedBy.String(),
		DueDate:    dueDate,
		CreatedAt:  issue.CreatedAt.Time,
		UpdatedAt:  issue.UpdatedAt.Time,
	}
}
