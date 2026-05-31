package workspaces

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
	ErrWorkspaceNotFound  = common.NewAppError(http.StatusNotFound, "workspace not found")
	ErrInvalidOwnerID     = common.NewAppError(http.StatusBadRequest, "invalid owner ID")
	ErrInvalidMemberID    = common.NewAppError(http.StatusBadRequest, "invalid member ID")
	ErrNotWorkspaceMember = common.NewAppError(http.StatusForbidden, "user is not a member of the workspace")
)

type WorkspaceStore interface {
	CreateWorkspace(ctx context.Context, arg db.CreateWorkspaceParams) (db.Workspace, error)
	GetWorkspaceByID(ctx context.Context, id pgtype.UUID) (db.Workspace, error)
	GetWorkspacesByMemberID(ctx context.Context, memberID pgtype.UUID) ([]db.Workspace, error)
	UpdateWorkspace(ctx context.Context, arg db.UpdateWorkspaceParams) (db.Workspace, error)
	DeleteWorkspace(ctx context.Context, id pgtype.UUID) error
	AddMember(ctx context.Context, arg db.AddMemberParams) (db.WorkspaceMember, error)
	GetMember(ctx context.Context, arg db.GetMemberParams) (db.WorkspaceMember, error)
}

type CreateWorkspaceInput struct {
	Name    string
	OwnerID string
}

type GetWorkspaceByIdInput struct {
	ID       string
	MemberID string
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
	if err != nil {
		return Workspace{}, err
	}
	_, err = w.store.AddMember(ctx, db.AddMemberParams{
		WorkspaceID: workspace.ID,
		UserID:      ownerID,
		Role:        "administrator",
	})
	if err != nil {
		return Workspace{}, err
	}
	return mapDbWorkspaceToDomain(workspace), nil
}

func (w *WorkspaceService) GetWorkspaceByID(ctx context.Context, input GetWorkspaceByIdInput) (Workspace, error) {
	var workspaceID pgtype.UUID
	if err := workspaceID.Scan(input.ID); err != nil {
		return Workspace{}, common.ErrInvalidWorkspaceID
	}
	var memberID pgtype.UUID
	if err := memberID.Scan(input.MemberID); err != nil {
		return Workspace{}, ErrInvalidMemberID
	}
	workspace, err := w.store.GetWorkspaceByID(ctx, workspaceID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Workspace{}, ErrWorkspaceNotFound
	}
	if err != nil {
		return Workspace{}, err
	}
	_, err = w.store.GetMember(ctx, db.GetMemberParams{
		WorkspaceID: workspaceID,
		UserID:      memberID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Workspace{}, ErrNotWorkspaceMember
	}
	if err != nil {
		return Workspace{}, err
	}
	return mapDbWorkspaceToDomain(workspace), nil
}

func (w *WorkspaceService) GetWorkspacesByMemberID(ctx context.Context, memberID string) ([]Workspace, error) {
	var memberUUID pgtype.UUID
	if err := memberUUID.Scan(memberID); err != nil {
		return nil, ErrInvalidMemberID
	}
	workspaces, err := w.store.GetWorkspacesByMemberID(ctx, memberUUID)
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
