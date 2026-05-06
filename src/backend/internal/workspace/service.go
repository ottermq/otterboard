package workspace

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
	ErrWorkspaceNotFound = common.NewAppError(http.StatusNotFound, "workspace not found")
	ErrInvalidOwnerID    = common.NewAppError(http.StatusBadRequest, "invalid owner ID")
)

type WorkspaceStore interface {
	CreateWorkspace(ctx context.Context, arg db.CreateWorkspaceParams) (db.Workspace, error)
	GetWorkspaceByID(ctx context.Context, id pgtype.UUID) (db.Workspace, error)
	GetWorkspacesByOwnerID(ctx context.Context, ownerID pgtype.UUID) ([]db.Workspace, error)
	UpdateWorkspace(ctx context.Context, arg db.UpdateWorkspaceParams) (db.Workspace, error)
	DeleteWorkspace(ctx context.Context, id pgtype.UUID) error
}

type CreateWorkspaceInput struct {
	Name    string
	OwnerID string
}

type UpdateWorkspaceInput struct {
	ID          string
	Name        string
	RequestorID string
}

type DeleteWorkspaceInput struct {
	ID          string
	RequestorID string
}

type WorkspaceService struct {
	store WorkspaceStore
}

type Workspace struct {
	ID        string
	Name      string
	OwnerID   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewWorkspaceService(store WorkspaceStore) *WorkspaceService {
	return &WorkspaceService{
		store: store,
	}
}

func (w *WorkspaceService) CreateWorkspace(ctx context.Context, input CreateWorkspaceInput) (Workspace, error) {
	var ownerID pgtype.UUID
	if err := ownerID.Scan(input.OwnerID); err != nil {
		return Workspace{}, ErrInvalidOwnerID
	}
	workspace, err := w.store.CreateWorkspace(ctx, db.CreateWorkspaceParams{
		Name:    input.Name,
		OwnerID: ownerID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Workspace{}, ErrWorkspaceNotFound
	}
	if err != nil {
		return Workspace{}, err
	}
	return mapDbWorkspaceToDomain(workspace), nil
}

func (w *WorkspaceService) GetWorkspaceByID(ctx context.Context, id string) (Workspace, error) {
	var workspaceID pgtype.UUID
	if err := workspaceID.Scan(id); err != nil {
		return Workspace{}, common.ErrInvalidWorkspaceID
	}
	workspace, err := w.store.GetWorkspaceByID(ctx, workspaceID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Workspace{}, ErrWorkspaceNotFound
	}
	if err != nil {
		return Workspace{}, err
	}
	return mapDbWorkspaceToDomain(workspace), nil
}

func (w *WorkspaceService) GetWorkspacesByOwnerID(ctx context.Context, ownerID string) ([]Workspace, error) {
	var ownerUUID pgtype.UUID
	if err := ownerUUID.Scan(ownerID); err != nil {
		return nil, ErrInvalidOwnerID
	}
	workspaces, err := w.store.GetWorkspacesByOwnerID(ctx, ownerUUID)
	if err != nil {
		return nil, err
	}
	domainWorkspaces := make([]Workspace, len(workspaces))
	for i, ws := range workspaces {
		domainWorkspaces[i] = mapDbWorkspaceToDomain(ws)
	}
	return domainWorkspaces, nil
}

func (w *WorkspaceService) UpdateWorkspace(ctx context.Context, input UpdateWorkspaceInput) (Workspace, error) {
	var workspaceID pgtype.UUID
	if err := workspaceID.Scan(input.ID); err != nil {
		return Workspace{}, common.ErrInvalidWorkspaceID
	}
	var requestorID pgtype.UUID
	if err := requestorID.Scan(input.RequestorID); err != nil {
		return Workspace{}, common.ErrInvalidRequestorID
	}

	workspace, err := w.store.GetWorkspaceByID(ctx, workspaceID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Workspace{}, ErrWorkspaceNotFound
	}
	if err != nil {
		return Workspace{}, err
	}

	if workspace.OwnerID != requestorID {
		return Workspace{}, common.ErrForbidden
	}

	updatedWorkspace, err := w.store.UpdateWorkspace(ctx, db.UpdateWorkspaceParams{
		ID:   workspaceID,
		Name: input.Name,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Workspace{}, ErrWorkspaceNotFound
	}
	if err != nil {
		return Workspace{}, err
	}
	return mapDbWorkspaceToDomain(updatedWorkspace), nil
}

func (w *WorkspaceService) DeleteWorkspace(ctx context.Context, input DeleteWorkspaceInput) error {
	var workspaceID pgtype.UUID
	if err := workspaceID.Scan(input.ID); err != nil {
		return common.ErrInvalidWorkspaceID
	}
	var requestorID pgtype.UUID
	if err := requestorID.Scan(input.RequestorID); err != nil {
		return common.ErrInvalidRequestorID
	}

	workspace, err := w.store.GetWorkspaceByID(ctx, workspaceID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrWorkspaceNotFound
	}
	if err != nil {
		return err
	}
	if workspace.OwnerID != requestorID {
		return common.ErrForbidden
	}

	err = w.store.DeleteWorkspace(ctx, workspaceID)
	if err != nil {
		return err
	}
	return nil
}

func mapDbWorkspaceToDomain(workspace db.Workspace) Workspace {
	return Workspace{
		ID:        workspace.ID.String(),
		Name:      workspace.Name,
		OwnerID:   workspace.OwnerID.String(),
		CreatedAt: workspace.CreatedAt.Time,
		UpdatedAt: workspace.UpdatedAt.Time,
	}
}
