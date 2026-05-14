package members

import (
	"context"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ottermq/otterboard/src/backend/internal/common"
	"github.com/ottermq/otterboard/src/backend/internal/db"
)

var (
	ErrAlreadyMember  = common.NewAppError(http.StatusConflict, "user is already a member of the workspace")
	ErrMemberNotFound = common.NewAppError(http.StatusNotFound, "member not found")
)

type MemberStore interface {
	AddMember(ctx context.Context, arg db.AddMemberParams) (db.WorkspaceMember, error)
	GetMember(ctx context.Context, arg db.GetMemberParams) (db.WorkspaceMember, error)
	ListMembers(ctx context.Context, workspaceID pgtype.UUID) ([]db.WorkspaceMember, error)
	UpdateMemberRole(ctx context.Context, arg db.UpdateMemberRoleParams) (db.WorkspaceMember, error)
	RemoveMember(ctx context.Context, arg db.RemoveMemberParams) error
}

type MemberService struct {
	store MemberStore
}

func NewMemberService(store MemberStore) *MemberService {
	return &MemberService{
		store: store,
	}
}

type AddMemberInput struct {
	WorkspaceID string
	UserID      string
	Role        string
}

type UpdateMemberRoleInput struct {
	WorkspaceID string
	UserID      string
	NewRole     string
}

func (s *MemberService) AddMember(ctx context.Context, input AddMemberInput) (db.WorkspaceMember, error) {
	var workspaceID pgtype.UUID
	if err := workspaceID.Scan(input.WorkspaceID); err != nil {
		return db.WorkspaceMember{}, common.ErrInvalidWorkspaceID
	}
	var userID pgtype.UUID
	if err := userID.Scan(input.UserID); err != nil {
		return db.WorkspaceMember{}, common.ErrInvalidUserID
	}
	_, err := s.store.GetMember(ctx, db.GetMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	})
	if err == nil {
		return db.WorkspaceMember{}, ErrAlreadyMember
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return db.WorkspaceMember{}, err
	}
	membership, err := s.store.AddMember(ctx, db.AddMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Role:        input.Role,
	})
	if err != nil {
		return db.WorkspaceMember{}, err
	}
	return membership, nil
}

func (s *MemberService) ListMembers(ctx context.Context, workspaceID string) ([]db.WorkspaceMember, error) {
	var workspaceUUID pgtype.UUID
	if err := workspaceUUID.Scan(workspaceID); err != nil {
		return nil, common.ErrInvalidWorkspaceID
	}
	return s.store.ListMembers(ctx, workspaceUUID)
}

func (s *MemberService) UpdateMemberRole(ctx context.Context, input UpdateMemberRoleInput) (db.WorkspaceMember, error) {
	var workspaceID pgtype.UUID
	if err := workspaceID.Scan(input.WorkspaceID); err != nil {
		return db.WorkspaceMember{}, common.ErrInvalidWorkspaceID
	}
	var userID pgtype.UUID
	if err := userID.Scan(input.UserID); err != nil {
		return db.WorkspaceMember{}, common.ErrInvalidUserID
	}

	if _, err := s.store.GetMember(ctx, db.GetMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	}); errors.Is(err, pgx.ErrNoRows) {
		return db.WorkspaceMember{}, ErrMemberNotFound
	} else if err != nil {
		return db.WorkspaceMember{}, err
	}

	membership, err := s.store.UpdateMemberRole(ctx, db.UpdateMemberRoleParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Role:        input.NewRole,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return db.WorkspaceMember{}, ErrMemberNotFound
	}
	if err != nil {
		return db.WorkspaceMember{}, err
	}
	return membership, nil
}

func (s *MemberService) RemoveMember(ctx context.Context, workspaceID string, userID string) error {
	var workspaceUUID pgtype.UUID
	if err := workspaceUUID.Scan(workspaceID); err != nil {
		return common.ErrInvalidWorkspaceID
	}
	var userUUID pgtype.UUID
	if err := userUUID.Scan(userID); err != nil {
		return common.ErrInvalidUserID
	}

	return s.store.RemoveMember(ctx, db.RemoveMemberParams{
		WorkspaceID: workspaceUUID,
		UserID:      userUUID,
	})
}
