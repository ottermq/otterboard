package invites

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ottermq/otterboard/src/backend/internal/common"
	"github.com/ottermq/otterboard/src/backend/internal/db"
)

var (
	ErrInviteNotFound = common.NewAppError(http.StatusNotFound, "invite not found")
	ErrInviteExpired  = common.NewAppError(http.StatusBadRequest, "invite has expired")
	ErrInviteUsed     = common.NewAppError(http.StatusBadRequest, "invite has already been used")
)

type InviteStore interface {
	CreateInvite(ctx context.Context, arg db.CreateInviteParams) (db.Invite, error)
	GetInviteByToken(ctx context.Context, token string) (db.Invite, error)
	UseInvite(ctx context.Context, token string) (db.Invite, error)
	AddMember(ctx context.Context, arg db.AddMemberParams) (db.WorkspaceMember, error)
}

type InviteService struct {
	inviteStore InviteStore
}

type GenerateInviteInput struct {
	WorkspaceID string
	CreatedBy   string
	ExpiresIn   time.Duration
}

type AcceptInviteInput struct {
	Token  string
	UserID string
}

type Invite struct {
	ID          string
	WorkspaceID string
	CreatedBy   string
	Token       string
	CreatedAt   time.Time
	ExpiresAt   time.Time
	UsedAt      time.Time
}

func NewInviteService(store InviteStore) *InviteService {
	return &InviteService{
		inviteStore: store,
	}
}

func (s *InviteService) GenerateInvite(ctx context.Context, input GenerateInviteInput) (Invite, error) {
	var workspaceID pgtype.UUID
	if err := workspaceID.Scan(input.WorkspaceID); err != nil {
		return Invite{}, common.ErrInvalidWorkspaceID
	}
	var createdBy pgtype.UUID
	if err := createdBy.Scan(input.CreatedBy); err != nil {
		return Invite{}, common.ErrInvalidUserID
	}
	token, err := generateToken()
	if err != nil {
		return Invite{}, err
	}
	expiresAt := pgtype.Timestamptz{
		Time:  time.Now().Add(input.ExpiresIn),
		Valid: true,
	}
	invite, err := s.inviteStore.CreateInvite(ctx, db.CreateInviteParams{
		WorkspaceID: workspaceID,
		Token:       token,
		CreatedBy:   createdBy,
		ExpiresAt:   expiresAt,
	})
	if err != nil {
		return Invite{}, err
	}
	return mapDbInviteToDomain(invite), nil
}

func (s *InviteService) GetInvite(ctx context.Context, token string) (Invite, error) {
	invite, err := s.inviteStore.GetInviteByToken(ctx, token)
	if errors.Is(err, pgx.ErrNoRows) {
		return Invite{}, ErrInviteNotFound
	}
	if err != nil {
		return Invite{}, err
	}
	if !invite.ExpiresAt.Valid || invite.ExpiresAt.Time.Before(time.Now()) {
		return Invite{}, ErrInviteExpired
	}
	if invite.UsedAt.Valid {
		return Invite{}, ErrInviteUsed
	}
	return mapDbInviteToDomain(invite), nil
}

func (s *InviteService) AcceptInvite(ctx context.Context, input AcceptInviteInput) error {
	invite, err := s.GetInvite(ctx, input.Token)
	if err != nil {
		return err
	}
	var userID pgtype.UUID
	if err := userID.Scan(input.UserID); err != nil {
		return common.ErrInvalidUserID
	}
	var workspaceID pgtype.UUID
	if err := workspaceID.Scan(invite.WorkspaceID); err != nil {
		return common.ErrInvalidWorkspaceID
	}
	if _, err := s.inviteStore.AddMember(ctx, db.AddMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
		Role:        "member",
	}); err != nil {
		return err
	}
	_, err = s.inviteStore.UseInvite(ctx, input.Token)
	if err != nil {
		return err
	}
	return nil
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func mapDbInviteToDomain(invite db.Invite) Invite {
	var usedAt time.Time
	if invite.UsedAt.Valid {
		usedAt = invite.UsedAt.Time
	}
	return Invite{
		ID:          invite.ID.String(),
		WorkspaceID: invite.WorkspaceID.String(),
		CreatedBy:   invite.CreatedBy.String(),
		Token:       invite.Token,
		CreatedAt:   invite.CreatedAt.Time,
		ExpiresAt:   invite.ExpiresAt.Time,
		UsedAt:      usedAt,
	}
}
