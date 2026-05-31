package invites_test

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ottermq/otterboard/src/backend/internal/common"
	"github.com/ottermq/otterboard/src/backend/internal/db"
	"github.com/ottermq/otterboard/src/backend/internal/invites"
	"github.com/stretchr/testify/require"
)

type mockInviteStore struct {
	createInviteFn     func(ctx context.Context, arg db.CreateInviteParams) (db.Invite, error)
	getInviteByTokenFn func(ctx context.Context, token string) (db.Invite, error)
	useInviteFn        func(ctx context.Context, token string) (db.Invite, error)
	addMemberFn        func(ctx context.Context, arg db.AddMemberParams) (db.WorkspaceMember, error)
}

func (m *mockInviteStore) CreateInvite(ctx context.Context, arg db.CreateInviteParams) (db.Invite, error) {
	return m.createInviteFn(ctx, arg)
}

func (m *mockInviteStore) GetInviteByToken(ctx context.Context, token string) (db.Invite, error) {
	return m.getInviteByTokenFn(ctx, token)
}

func (m *mockInviteStore) UseInvite(ctx context.Context, token string) (db.Invite, error) {
	return m.useInviteFn(ctx, token)
}

func (m *mockInviteStore) AddMember(ctx context.Context, arg db.AddMemberParams) (db.WorkspaceMember, error) {
	return m.addMemberFn(ctx, arg)
}

func mustUUID(t *testing.T, id string) pgtype.UUID {
	t.Helper()

	var uuid pgtype.UUID
	require.NoError(t, uuid.Scan(id))
	return uuid
}

func TestGenerateInvite_Success(t *testing.T) {
	workspaceID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	createdBy := mustUUID(t, "22222222-2222-2222-2222-222222222222")
	expiresIn := 48 * time.Hour
	before := time.Now().Add(expiresIn)

	createCalled := false
	store := &mockInviteStore{
		createInviteFn: func(_ context.Context, arg db.CreateInviteParams) (db.Invite, error) {
			createCalled = true
			require.Equal(t, workspaceID, arg.WorkspaceID)
			require.Equal(t, createdBy, arg.CreatedBy)
			require.True(t, arg.ExpiresAt.Valid)
			require.WithinDuration(t, before, arg.ExpiresAt.Time, time.Second)

			decodedToken, err := base64.URLEncoding.DecodeString(arg.Token)
			require.NoError(t, err)
			require.Len(t, decodedToken, 32)

			return db.Invite{
				WorkspaceID: arg.WorkspaceID,
				CreatedBy:   arg.CreatedBy,
				Token:       arg.Token,
				ExpiresAt:   arg.ExpiresAt,
			}, nil
		},
	}

	service := invites.NewInviteService(store)
	invite, err := service.GenerateInvite(context.Background(), invites.GenerateInviteInput{
		WorkspaceID: workspaceID.String(),
		CreatedBy:   createdBy.String(),
		ExpiresIn:   expiresIn,
	})

	require.NoError(t, err)
	require.True(t, createCalled)
	require.Equal(t, workspaceID.String(), invite.WorkspaceID)
	require.Equal(t, createdBy.String(), invite.CreatedBy)
	require.NotEmpty(t, invite.Token)
	require.WithinDuration(t, before, invite.ExpiresAt, time.Second)
}

func TestGenerateInvite_InvalidIDs(t *testing.T) {
	store := &mockInviteStore{
		createInviteFn: func(_ context.Context, _ db.CreateInviteParams) (db.Invite, error) {
			t.Fatal("CreateInvite should not be called with invalid IDs")
			return db.Invite{}, nil
		},
	}

	service := invites.NewInviteService(store)
	_, err := service.GenerateInvite(context.Background(), invites.GenerateInviteInput{
		WorkspaceID: "not-a-uuid",
		CreatedBy:   "22222222-2222-2222-2222-222222222222",
		ExpiresIn:   24 * time.Hour,
	})
	require.ErrorIs(t, err, common.ErrInvalidWorkspaceID)

	_, err = service.GenerateInvite(context.Background(), invites.GenerateInviteInput{
		WorkspaceID: "11111111-1111-1111-1111-111111111111",
		CreatedBy:   "not-a-uuid",
		ExpiresIn:   24 * time.Hour,
	})
	require.ErrorIs(t, err, common.ErrInvalidUserID)
}

func TestGenerateInvite_StoreError(t *testing.T) {
	expectedErr := errors.New("store error")
	store := &mockInviteStore{
		createInviteFn: func(_ context.Context, _ db.CreateInviteParams) (db.Invite, error) {
			return db.Invite{}, expectedErr
		},
	}

	service := invites.NewInviteService(store)
	_, err := service.GenerateInvite(context.Background(), invites.GenerateInviteInput{
		WorkspaceID: "11111111-1111-1111-1111-111111111111",
		CreatedBy:   "22222222-2222-2222-2222-222222222222",
		ExpiresIn:   24 * time.Hour,
	})

	require.ErrorIs(t, err, expectedErr)
}

func TestGetInvite_Success(t *testing.T) {
	expectedInvite := db.Invite{
		WorkspaceID: mustUUID(t, "11111111-1111-1111-1111-111111111111"),
		CreatedBy:   mustUUID(t, "22222222-2222-2222-2222-222222222222"),
		Token:       "valid-token",
		ExpiresAt:   pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
	}
	store := &mockInviteStore{
		getInviteByTokenFn: func(_ context.Context, token string) (db.Invite, error) {
			require.Equal(t, "valid-token", token)
			return expectedInvite, nil
		},
	}

	service := invites.NewInviteService(store)
	invite, err := service.GetInvite(context.Background(), "valid-token")

	require.NoError(t, err)
	require.Equal(t, expectedInvite.WorkspaceID.String(), invite.WorkspaceID)
	require.Equal(t, expectedInvite.CreatedBy.String(), invite.CreatedBy)
	require.Equal(t, expectedInvite.Token, invite.Token)
	require.Equal(t, expectedInvite.ExpiresAt.Time, invite.ExpiresAt)
}

func TestGetInvite_NotFound(t *testing.T) {
	store := &mockInviteStore{
		getInviteByTokenFn: func(_ context.Context, _ string) (db.Invite, error) {
			return db.Invite{}, pgx.ErrNoRows
		},
	}

	service := invites.NewInviteService(store)
	_, err := service.GetInvite(context.Background(), "missing-token")

	require.ErrorIs(t, err, invites.ErrInviteNotFound)
}

func TestGetInvite_StoreError(t *testing.T) {
	expectedErr := errors.New("store error")
	store := &mockInviteStore{
		getInviteByTokenFn: func(_ context.Context, _ string) (db.Invite, error) {
			return db.Invite{}, expectedErr
		},
	}

	service := invites.NewInviteService(store)
	_, err := service.GetInvite(context.Background(), "valid-token")

	require.ErrorIs(t, err, expectedErr)
}

func TestGetInvite_Expired(t *testing.T) {
	store := &mockInviteStore{
		getInviteByTokenFn: func(_ context.Context, _ string) (db.Invite, error) {
			return db.Invite{
				Token:     "expired-token",
				ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(-time.Minute), Valid: true},
			}, nil
		},
	}

	service := invites.NewInviteService(store)
	_, err := service.GetInvite(context.Background(), "expired-token")

	require.ErrorIs(t, err, invites.ErrInviteExpired)
}

func TestGetInvite_InvalidExpiresAt(t *testing.T) {
	store := &mockInviteStore{
		getInviteByTokenFn: func(_ context.Context, _ string) (db.Invite, error) {
			return db.Invite{
				Token:     "invalid-expires-at-token",
				ExpiresAt: pgtype.Timestamptz{Valid: false},
			}, nil
		},
	}

	service := invites.NewInviteService(store)
	_, err := service.GetInvite(context.Background(), "invalid-expires-at-token")

	require.ErrorIs(t, err, invites.ErrInviteExpired)
}

func TestGetInvite_Used(t *testing.T) {
	store := &mockInviteStore{
		getInviteByTokenFn: func(_ context.Context, _ string) (db.Invite, error) {
			return db.Invite{
				Token:     "used-token",
				ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
				UsedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}

	service := invites.NewInviteService(store)
	_, err := service.GetInvite(context.Background(), "used-token")

	require.ErrorIs(t, err, invites.ErrInviteUsed)
}

func TestAcceptInvite_Success(t *testing.T) {
	workspaceID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	userID := mustUUID(t, "22222222-2222-2222-2222-222222222222")

	addCalled := false
	useCalled := false
	store := &mockInviteStore{
		getInviteByTokenFn: func(_ context.Context, token string) (db.Invite, error) {
			require.Equal(t, "valid-token", token)
			return db.Invite{
				WorkspaceID: workspaceID,
				Token:       token,
				ExpiresAt:   pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
			}, nil
		},
		addMemberFn: func(_ context.Context, arg db.AddMemberParams) (db.WorkspaceMember, error) {
			addCalled = true
			require.Equal(t, workspaceID, arg.WorkspaceID)
			require.Equal(t, userID, arg.UserID)
			require.Equal(t, "member", arg.Role)
			return db.WorkspaceMember{
				WorkspaceID: arg.WorkspaceID,
				UserID:      arg.UserID,
				Role:        arg.Role,
			}, nil
		},
		useInviteFn: func(_ context.Context, token string) (db.Invite, error) {
			useCalled = true
			require.Equal(t, "valid-token", token)
			return db.Invite{Token: token}, nil
		},
	}

	service := invites.NewInviteService(store)
	err := service.AcceptInvite(context.Background(), invites.AcceptInviteInput{
		Token:  "valid-token",
		UserID: userID.String(),
	})

	require.NoError(t, err)
	require.True(t, addCalled)
	require.True(t, useCalled)
}

func TestAcceptInvite_GetInviteError(t *testing.T) {
	store := &mockInviteStore{
		getInviteByTokenFn: func(_ context.Context, _ string) (db.Invite, error) {
			return db.Invite{}, pgx.ErrNoRows
		},
		addMemberFn: func(_ context.Context, _ db.AddMemberParams) (db.WorkspaceMember, error) {
			t.Fatal("AddMember should not be called when invite lookup fails")
			return db.WorkspaceMember{}, nil
		},
		useInviteFn: func(_ context.Context, _ string) (db.Invite, error) {
			t.Fatal("UseInvite should not be called when invite lookup fails")
			return db.Invite{}, nil
		},
	}

	service := invites.NewInviteService(store)
	err := service.AcceptInvite(context.Background(), invites.AcceptInviteInput{
		Token:  "missing-token",
		UserID: "22222222-2222-2222-2222-222222222222",
	})

	require.ErrorIs(t, err, invites.ErrInviteNotFound)
}

func TestAcceptInvite_InvalidUserID(t *testing.T) {
	store := &mockInviteStore{
		getInviteByTokenFn: func(_ context.Context, _ string) (db.Invite, error) {
			return db.Invite{
				WorkspaceID: mustUUID(t, "11111111-1111-1111-1111-111111111111"),
				Token:       "valid-token",
				ExpiresAt:   pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
			}, nil
		},
		addMemberFn: func(_ context.Context, _ db.AddMemberParams) (db.WorkspaceMember, error) {
			t.Fatal("AddMember should not be called with an invalid user ID")
			return db.WorkspaceMember{}, nil
		},
		useInviteFn: func(_ context.Context, _ string) (db.Invite, error) {
			t.Fatal("UseInvite should not be called with an invalid user ID")
			return db.Invite{}, nil
		},
	}

	service := invites.NewInviteService(store)
	err := service.AcceptInvite(context.Background(), invites.AcceptInviteInput{
		Token:  "valid-token",
		UserID: "not-a-uuid",
	})

	require.ErrorIs(t, err, common.ErrInvalidUserID)
}

func TestAcceptInvite_AddMemberError(t *testing.T) {
	expectedErr := errors.New("add member error")
	store := &mockInviteStore{
		getInviteByTokenFn: func(_ context.Context, _ string) (db.Invite, error) {
			return db.Invite{
				WorkspaceID: mustUUID(t, "11111111-1111-1111-1111-111111111111"),
				Token:       "valid-token",
				ExpiresAt:   pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
			}, nil
		},
		addMemberFn: func(_ context.Context, _ db.AddMemberParams) (db.WorkspaceMember, error) {
			return db.WorkspaceMember{}, expectedErr
		},
		useInviteFn: func(_ context.Context, _ string) (db.Invite, error) {
			t.Fatal("UseInvite should not be called when adding the member fails")
			return db.Invite{}, nil
		},
	}

	service := invites.NewInviteService(store)
	err := service.AcceptInvite(context.Background(), invites.AcceptInviteInput{
		Token:  "valid-token",
		UserID: "22222222-2222-2222-2222-222222222222",
	})

	require.ErrorIs(t, err, expectedErr)
}

func TestAcceptInvite_UseInviteError(t *testing.T) {
	expectedErr := errors.New("use invite error")
	store := &mockInviteStore{
		getInviteByTokenFn: func(_ context.Context, _ string) (db.Invite, error) {
			return db.Invite{
				WorkspaceID: mustUUID(t, "11111111-1111-1111-1111-111111111111"),
				Token:       "valid-token",
				ExpiresAt:   pgtype.Timestamptz{Time: time.Now().Add(time.Hour), Valid: true},
			}, nil
		},
		addMemberFn: func(_ context.Context, arg db.AddMemberParams) (db.WorkspaceMember, error) {
			return db.WorkspaceMember{
				WorkspaceID: arg.WorkspaceID,
				UserID:      arg.UserID,
				Role:        arg.Role,
			}, nil
		},
		useInviteFn: func(_ context.Context, _ string) (db.Invite, error) {
			return db.Invite{}, expectedErr
		},
	}

	service := invites.NewInviteService(store)
	err := service.AcceptInvite(context.Background(), invites.AcceptInviteInput{
		Token:  "valid-token",
		UserID: "22222222-2222-2222-2222-222222222222",
	})

	require.ErrorIs(t, err, expectedErr)
}
