package issues

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ottermq/otterboard/src/backend/internal/common"
	"github.com/ottermq/otterboard/src/backend/internal/db"
)

var (
	ErrInvalidTitle      = common.NewAppError(http.StatusBadRequest, "invalid issue title")
	ErrInvalidType       = common.NewAppError(http.StatusBadRequest, "invalid issue type")
	ErrInvalidStatus     = common.NewAppError(http.StatusBadRequest, "invalid issue status")
	ErrInvalidAssigneeID = common.NewAppError(http.StatusBadRequest, "invalid assignee ID")
	ErrInvalidIssueID    = common.NewAppError(http.StatusBadRequest, "invalid issue ID")
	ErrIssueNotFound     = common.NewAppError(http.StatusNotFound, "issue not found")
)

const (
	DefaultIssueStatus = "backlog"
	DefaultLimit       = 20
	MaxLimit           = 100
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

var validSortFields = map[string]struct{}{
	"title":      {},
	"type":       {},
	"status":     {},
	"due_date":   {},
	"created_at": {},
}

type IssueStore interface {
	GetMaxPositionByProjectAndStatus(ctx context.Context, arg db.GetMaxPositionByProjectAndStatusParams) (any, error)
	CreateIssue(ctx context.Context, arg db.CreateIssueParams) (db.Issue, error)
	GetIssueByID(ctx context.Context, arg db.GetIssueByIDParams) (db.Issue, error)
	ListIssuesByProjectFiltered(ctx context.Context, arg db.ListIssuesByProjectFilteredParams) ([]db.Issue, error)
	CountIssuesByProjectFiltered(ctx context.Context, arg db.CountIssuesByProjectFilteredParams) (int64, error)
	ListIssuesByWorkspaceFiltered(ctx context.Context, arg db.ListIssuesByWorkspaceFilteredParams) ([]db.Issue, error)
	CountIssuesByWorkspaceFiltered(ctx context.Context, arg db.CountIssuesByWorkspaceFilteredParams) (int64, error)
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

type GetIssueByIdInput struct {
	ID        string
	ProjectID string
}

type GetMaxPositionByProjectAndStatusInput struct {
	ProjectID string
	Status    string
}

type CommonFiltersInput struct {
	Status     string
	Type       string
	AssigneeID string
	DueBefore  *time.Time
	DueAfter   *time.Time
	SortBy     string
	SortOrder  string
}

type NormalizedFiltersInput struct {
	Status     pgtype.Text
	Type       pgtype.Text
	AssigneeID pgtype.UUID
	DueBefore  pgtype.Date
	DueAfter   pgtype.Date
	SortBy     string
	SortOrder  string
}

type ListIssuesByProjectInput struct {
	ProjectID string
	Filters   CommonFiltersInput
	Page      int32
	Limit     int32
}

type ListIssuesByWorkspaceInput struct {
	WorkspaceID     string
	ProjectIDFilter string
	Filters         CommonFiltersInput
	Page            int32
	Limit           int32
}

type CountIssuesByProjectFilteredInput struct {
	ProjectID string
	Filters   CommonFiltersInput
}

type CountIssuesByWorkspaceFilteredInput struct {
	WorkspaceID     string
	ProjectIDFilter string
	Filters         CommonFiltersInput
}

type UpdateIssueInput struct {
	ID         string
	ProjectID  string
	Title      string
	Overview   string
	Type       string
	Status     string
	Position   float64
	AssigneeID string
	DueDate    *time.Time
}

type DeleteIssueInput struct {
	ID        string
	ProjectID string
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

func (i *IssueService) GetIssueByID(ctx context.Context, input GetIssueByIdInput) (Issue, error) {
	var issueUUID pgtype.UUID
	if err := issueUUID.Scan(input.ID); err != nil {
		return Issue{}, ErrInvalidIssueID
	}

	var projectID pgtype.UUID
	if err := projectID.Scan(input.ProjectID); err != nil {
		return Issue{}, common.ErrInvalidProjectID
	}

	issue, err := i.store.GetIssueByID(ctx, db.GetIssueByIDParams{
		ID:        issueUUID,
		ProjectID: projectID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Issue{}, ErrIssueNotFound
	}
	if err != nil {
		return Issue{}, err
	}
	return mapToIssueDomain(issue), nil
}

func (i *IssueService) ListIssuesByProjectFiltered(ctx context.Context, input ListIssuesByProjectInput) ([]Issue, error) {
	var projectUUID pgtype.UUID
	if err := projectUUID.Scan(input.ProjectID); err != nil {
		return []Issue{}, common.ErrInvalidProjectID
	}

	normalized, err := normalizeFilters(input.Filters)
	if err != nil {
		return []Issue{}, err
	}

	page := max(input.Page, 1)
	limit := input.Limit
	if limit < 1 {
		limit = DefaultLimit
	} else if limit > MaxLimit {
		limit = MaxLimit
	}
	offset := (page - 1) * limit

	issues, err := i.store.ListIssuesByProjectFiltered(ctx, db.ListIssuesByProjectFilteredParams{
		ProjectID:  projectUUID,
		Status:     normalized.Status,
		Type:       normalized.Type,
		AssigneeID: normalized.AssigneeID,
		DueBefore:  normalized.DueBefore,
		DueAfter:   normalized.DueAfter,
		SortBy:     normalized.SortBy,
		SortOrder:  normalized.SortOrder,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return []Issue{}, err
	}
	domainIssues := make([]Issue, len(issues))
	for idx, issue := range issues {
		domainIssues[idx] = mapToIssueDomain(issue)
	}
	return domainIssues, nil
}

func (i *IssueService) CountIssuesByProjectFiltered(ctx context.Context, input CountIssuesByProjectFilteredInput) (int64, error) {
	var projectUUID pgtype.UUID
	if err := projectUUID.Scan(input.ProjectID); err != nil {
		return 0, common.ErrInvalidProjectID
	}

	normalized, err := normalizeFilters(input.Filters)
	if err != nil {
		return 0, err
	}

	count, err := i.store.CountIssuesByProjectFiltered(ctx, db.CountIssuesByProjectFilteredParams{
		ProjectID:  projectUUID,
		Status:     normalized.Status,
		Type:       normalized.Type,
		AssigneeID: normalized.AssigneeID,
		DueBefore:  normalized.DueBefore,
		DueAfter:   normalized.DueAfter,
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (i *IssueService) ListIssuesByWorkspaceFiltered(ctx context.Context, input ListIssuesByWorkspaceInput) ([]Issue, error) {
	var workspaceUUID pgtype.UUID
	if err := workspaceUUID.Scan(input.WorkspaceID); err != nil {
		return []Issue{}, common.ErrInvalidWorkspaceID
	}
	var projectUUID pgtype.UUID
	if input.ProjectIDFilter == "" {
		projectUUID = pgtype.UUID{Valid: false}
	} else if err := projectUUID.Scan(input.ProjectIDFilter); err != nil {
		return []Issue{}, common.ErrInvalidProjectID
	}

	normalized, err := normalizeFilters(input.Filters)
	if err != nil {
		return []Issue{}, err
	}

	page := max(input.Page, 1)
	limit := input.Limit
	if limit < 1 {
		limit = DefaultLimit
	} else if limit > MaxLimit {
		limit = MaxLimit
	}
	offset := (page - 1) * limit
	dbIssues, err := i.store.ListIssuesByWorkspaceFiltered(ctx, db.ListIssuesByWorkspaceFilteredParams{
		WorkspaceID:     workspaceUUID,
		ProjectIDFilter: projectUUID,
		Status:          normalized.Status,
		Type:            normalized.Type,
		AssigneeID:      normalized.AssigneeID,
		DueBefore:       normalized.DueBefore,
		DueAfter:        normalized.DueAfter,
		SortBy:          normalized.SortBy,
		SortOrder:       normalized.SortOrder,
		Limit:           limit,
		Offset:          offset,
	})
	if err != nil {
		return []Issue{}, err
	}
	domainIssues := make([]Issue, len(dbIssues))
	for idx, issue := range dbIssues {
		domainIssues[idx] = mapToIssueDomain(issue)
	}
	return domainIssues, nil
}

func (i *IssueService) CountIssuesByWorkspaceFiltered(ctx context.Context, input CountIssuesByWorkspaceFilteredInput) (int64, error) {
	var workspaceUUID pgtype.UUID
	if err := workspaceUUID.Scan(input.WorkspaceID); err != nil {
		return 0, common.ErrInvalidWorkspaceID
	}

	var projectUUID pgtype.UUID
	if input.ProjectIDFilter == "" {
		projectUUID = pgtype.UUID{Valid: false}
	} else if err := projectUUID.Scan(input.ProjectIDFilter); err != nil {
		return 0, common.ErrInvalidProjectID
	}

	normalized, err := normalizeFilters(input.Filters)
	if err != nil {
		return 0, err
	}

	count, err := i.store.CountIssuesByWorkspaceFiltered(ctx, db.CountIssuesByWorkspaceFilteredParams{
		WorkspaceID:     workspaceUUID,
		ProjectIDFilter: projectUUID,
		Status:          normalized.Status,
		Type:            normalized.Type,
		AssigneeID:      normalized.AssigneeID,
		DueBefore:       normalized.DueBefore,
		DueAfter:        normalized.DueAfter,
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (i *IssueService) UpdateIssue(ctx context.Context, input UpdateIssueInput) (Issue, error) {
	var issueUUID pgtype.UUID
	if err := issueUUID.Scan(input.ID); err != nil {
		return Issue{}, ErrInvalidIssueID
	}
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
	if !IsValidStatus(input.Status) {
		return Issue{}, ErrInvalidStatus
	}
	position := max(input.Position, 0)

	issue, err := i.store.UpdateIssue(ctx, db.UpdateIssueParams{
		ID:         issueUUID,
		ProjectID:  projectUUID,
		Title:      input.Title,
		Overview:   overview,
		Type:       input.Type,
		Status:     input.Status,
		Position:   position,
		AssigneeID: assigneeUUID,
		DueDate:    dueDateToPgDate(input.DueDate),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Issue{}, ErrIssueNotFound
	}
	if err != nil {
		return Issue{}, err
	}
	return mapToIssueDomain(issue), nil
}

func (i *IssueService) DeleteIssue(ctx context.Context, input DeleteIssueInput) error {
	var issueUUID pgtype.UUID
	if err := issueUUID.Scan(input.ID); err != nil {
		return ErrInvalidIssueID
	}
	var projectUUID pgtype.UUID
	if err := projectUUID.Scan(input.ProjectID); err != nil {
		return common.ErrInvalidProjectID
	}

	err := i.store.DeleteIssue(ctx, db.DeleteIssueParams{
		ID:        issueUUID,
		ProjectID: projectUUID,
	})
	if err != nil {
		return err
	}
	return nil
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

func normalizeFilters(input CommonFiltersInput) (NormalizedFiltersInput, error) {
	var normalized NormalizedFiltersInput
	if input.AssigneeID != "" {
		if err := normalized.AssigneeID.Scan(input.AssigneeID); err != nil {
			return normalized, ErrInvalidAssigneeID
		}
	}

	if _, ok := validStatuses[input.Status]; ok {
		normalized.Status = pgtype.Text{String: input.Status, Valid: true}
	} else {
		normalized.Status = pgtype.Text{Valid: false}
	}

	if _, ok := validTypes[input.Type]; ok {
		normalized.Type = pgtype.Text{String: input.Type, Valid: true}
	} else {
		normalized.Type = pgtype.Text{Valid: false}
	}

	normalized.DueBefore = dueDateToPgDate(input.DueBefore)

	normalized.DueAfter = dueDateToPgDate(input.DueAfter)

	normalized.SortBy = input.SortBy
	if _, ok := validSortFields[normalized.SortBy]; !ok {
		normalized.SortBy = ""
	}

	normalized.SortOrder = input.SortOrder
	if normalized.SortOrder != "desc" {
		normalized.SortOrder = "asc"
	}
	return normalized, nil
}
