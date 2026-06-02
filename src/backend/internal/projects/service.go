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
	if err := projectID.Scan(input.WorkspaceID); err != nil {
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
