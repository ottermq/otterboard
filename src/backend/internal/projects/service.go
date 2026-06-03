package projects

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

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

var (
	ErrProjectNotFound    = common.NewAppError(http.StatusNotFound, "project not found")
	ErrInvalidProjectID   = common.NewAppError(http.StatusBadRequest, "invalid project ID")
	ErrInvalidProjectName = common.NewAppError(http.StatusBadRequest, "project name is required")
)

type Project struct {
	ID          string
	WorkspaceID string
	Name        string
	ImageUrl    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ProjectStore interface {
	CreateProject(ctx context.Context, arg db.CreateProjectParams) (db.Project, error)
	GetProjectByID(ctx context.Context, arg db.GetProjectByIDParams) (db.Project, error)
	ListProjectsByWorkspace(ctx context.Context, arg db.ListProjectsByWorkspaceParams) ([]db.Project, error)
	UpdateProject(ctx context.Context, arg db.UpdateProjectParams) (db.Project, error)
	DeleteProject(ctx context.Context, arg db.DeleteProjectParams) error
}

type ProjectService struct {
	store ProjectStore
}

func NewProjectService(store ProjectStore) *ProjectService {
	return &ProjectService{
		store: store,
	}
}

type CreateProjectInput struct {
	WorkspaceID string
	Name        string
	ImageUrl    string
}

type GetProjectByIdInput struct {
	ID          string
	WorkspaceID string
}

type ListProjectsByWorkspaceInput struct {
	WorkspaceID string
	Page        int32
	Limit       int32
}

type UpdateProjectInput struct {
	ID          string
	WorkspaceID string
	Name        string
	ImageUrl    string
}

type DeleteProjectInput struct {
	ID          string
	WorkspaceID string
}

func (p *ProjectService) CreateProject(ctx context.Context, input CreateProjectInput) (Project, error) {
	var workspaceID pgtype.UUID
	if err := workspaceID.Scan(input.WorkspaceID); err != nil {
		return Project{}, common.ErrInvalidWorkspaceID
	}

	if input.Name == "" {
		return Project{}, ErrInvalidProjectName
	}
	imageUrl := pgtype.Text{Valid: false}
	if input.ImageUrl != "" {
		imageUrl = pgtype.Text{String: input.ImageUrl, Valid: true}
	}

	project, err := p.store.CreateProject(ctx, db.CreateProjectParams{
		WorkspaceID: workspaceID,
		Name:        input.Name,
		ImageUrl:    imageUrl,
	})
	if err != nil {
		return Project{}, err
	}
	return mapDbProjectToDomain(project), nil
}

func (p *ProjectService) GetProjectByID(ctx context.Context, input GetProjectByIdInput) (Project, error) {
	var projectID pgtype.UUID
	if err := projectID.Scan(input.ID); err != nil {
		return Project{}, ErrInvalidProjectID
	}

	var workspaceID pgtype.UUID
	if err := workspaceID.Scan(input.WorkspaceID); err != nil {
		return Project{}, common.ErrInvalidWorkspaceID
	}

	project, err := p.store.GetProjectByID(ctx, db.GetProjectByIDParams{
		ID:          projectID,
		WorkspaceID: workspaceID,
	})

	if errors.Is(err, pgx.ErrNoRows) {
		return Project{}, ErrProjectNotFound
	}
	if err != nil {
		return Project{}, err
	}
	return mapDbProjectToDomain(project), nil
}

func (p *ProjectService) ListProjectsByWorkspace(ctx context.Context, input ListProjectsByWorkspaceInput) ([]Project, error) {
	var workspaceID pgtype.UUID
	if err := workspaceID.Scan(input.WorkspaceID); err != nil {
		return []Project{}, common.ErrInvalidWorkspaceID
	}

	page := input.Page
	if page < 1 {
		page = 1
	}
	limit := input.Limit
	if limit < 1 {
		limit = DefaultLimit
	} else if limit > MaxLimit {
		limit = MaxLimit
	}
	offset := (page - 1) * limit

	projects, err := p.store.ListProjectsByWorkspace(ctx, db.ListProjectsByWorkspaceParams{
		WorkspaceID: workspaceID,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		return []Project{}, err
	}
	domainProjects := make([]Project, len(projects))
	for i, proj := range projects {
		domainProjects[i] = mapDbProjectToDomain(proj)
	}
	return domainProjects, nil
}

func (p *ProjectService) UpdateProject(ctx context.Context, input UpdateProjectInput) (Project, error) {
	var workspaceID pgtype.UUID
	if err := workspaceID.Scan(input.WorkspaceID); err != nil {
		return Project{}, common.ErrInvalidWorkspaceID
	}

	var id pgtype.UUID
	if err := id.Scan(input.ID); err != nil {
		return Project{}, ErrInvalidProjectID
	}

	if input.Name == "" {
		return Project{}, ErrInvalidProjectName
	}

	var imageUrl pgtype.Text
	if input.ImageUrl == "" {
		imageUrl = pgtype.Text{Valid: false}
	} else {
		imageUrl = pgtype.Text{String: input.ImageUrl, Valid: true}
	}

	updated, err := p.store.UpdateProject(ctx, db.UpdateProjectParams{
		ID:          id,
		WorkspaceID: workspaceID,
		Name:        input.Name,
		ImageUrl:    imageUrl,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Project{}, ErrProjectNotFound
	}
	if err != nil {
		return Project{}, err
	}
	return mapDbProjectToDomain(updated), nil
}

func (p *ProjectService) DeleteProject(ctx context.Context, input DeleteProjectInput) error {
	var workspaceID pgtype.UUID
	if err := workspaceID.Scan(input.WorkspaceID); err != nil {
		return common.ErrInvalidWorkspaceID
	}

	var id pgtype.UUID
	if err := id.Scan(input.ID); err != nil {
		return ErrInvalidProjectID
	}

	err := p.store.DeleteProject(ctx, db.DeleteProjectParams{
		WorkspaceID: workspaceID,
		ID:          id,
	})

	return err
}

func mapDbProjectToDomain(project db.Project) Project {
	return Project{
		ID:          project.ID.String(),
		WorkspaceID: project.WorkspaceID.String(),
		Name:        project.Name,
		ImageUrl:    project.ImageUrl.String,
		CreatedAt:   project.CreatedAt.Time,
		UpdatedAt:   project.UpdatedAt.Time,
	}
}
