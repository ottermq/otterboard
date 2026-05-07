package members_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ottermq/otterboard/src/backend/internal/common"
	"github.com/ottermq/otterboard/src/backend/internal/db"
	"github.com/ottermq/otterboard/src/backend/internal/members"
	"github.com/stretchr/testify/require"
)

type mockMemberStore struct {
	addMemberFn        func(ctx context.Context, arg db.AddMemberParams) (db.WorkspaceMember, error)
	getMemberFn        func(ctx context.Context, arg db.GetMemberParams) (db.WorkspaceMember, error)
	listMembersFn      func(ctx context.Context, workspaceID pgtype.UUID) ([]db.WorkspaceMember, error)
	updateMemberRoleFn func(ctx context.Context, arg db.UpdateMemberRoleParams) (db.WorkspaceMember, error)
	removeMemberFn     func(ctx context.Context, arg db.RemoveMemberParams) error
}

func (m *mockMemberStore) AddMember(ctx context.Context, arg db.AddMemberParams) (db.WorkspaceMember, error) {
	return m.addMemberFn(ctx, arg)
}

func (m *mockMemberStore) GetMember(ctx context.Context, arg db.GetMemberParams) (db.WorkspaceMember, error) {
	return m.getMemberFn(ctx, arg)
}

func (m *mockMemberStore) ListMembers(ctx context.Context, workspaceID pgtype.UUID) ([]db.WorkspaceMember, error) {
	return m.listMembersFn(ctx, workspaceID)
}

func (m *mockMemberStore) UpdateMemberRole(ctx context.Context, arg db.UpdateMemberRoleParams) (db.WorkspaceMember, error) {
	return m.updateMemberRoleFn(ctx, arg)
}

func (m *mockMemberStore) RemoveMember(ctx context.Context, arg db.RemoveMemberParams) error {
	return m.removeMemberFn(ctx, arg)
}

func mustUUID(t *testing.T, id string) pgtype.UUID {
	t.Helper()

	var uuid pgtype.UUID
	require.NoError(t, uuid.Scan(id))
	return uuid
}

func TestAddMember_Success(t *testing.T) {
	store := &mockMemberStore{
		getMemberFn: func(_ context.Context, _ db.GetMemberParams) (db.WorkspaceMember, error) {
			return db.WorkspaceMember{}, pgx.ErrNoRows
		},
		addMemberFn: func(_ context.Context, arg db.AddMemberParams) (db.WorkspaceMember, error) {
			return db.WorkspaceMember{
				WorkspaceID: arg.WorkspaceID,
				UserID:      arg.UserID,
				Role:        arg.Role,
			}, nil
		},
	}

	service := members.NewMemberService(store)
	member, err := service.AddMember(context.Background(), members.AddMemberInput{
		WorkspaceID: "11111111-1111-1111-1111-111111111111",
		UserID:      "22222222-2222-2222-2222-222222222222",
		Role:        "member",
	})

	require.NoError(t, err)
	require.Equal(t, "11111111-1111-1111-1111-111111111111", member.WorkspaceID.String())
	require.Equal(t, "22222222-2222-2222-2222-222222222222", member.UserID.String())
	require.Equal(t, "member", member.Role)
}

func TestAddMember_InvalidIDs(t *testing.T) {
	service := members.NewMemberService(&mockMemberStore{})

	_, err := service.AddMember(context.Background(), members.AddMemberInput{
		WorkspaceID: "not-a-uuid",
		UserID:      "22222222-2222-2222-2222-222222222222",
		Role:        "member",
	})
	require.ErrorIs(t, err, common.ErrInvalidWorkspaceID)

	_, err = service.AddMember(context.Background(), members.AddMemberInput{
		WorkspaceID: "11111111-1111-1111-1111-111111111111",
		UserID:      "not-a-uuid",
		Role:        "member",
	})
	require.ErrorIs(t, err, common.ErrInvalidUserID)
}

func TestAddMember_AlreadyMember(t *testing.T) {
	store := &mockMemberStore{
		getMemberFn: func(_ context.Context, _ db.GetMemberParams) (db.WorkspaceMember, error) {
			return db.WorkspaceMember{}, nil
		},
		addMemberFn: func(_ context.Context, _ db.AddMemberParams) (db.WorkspaceMember, error) {
			t.Fatal("AddMember should not be called when membership already exists")
			return db.WorkspaceMember{}, nil
		},
	}

	service := members.NewMemberService(store)
	_, err := service.AddMember(context.Background(), members.AddMemberInput{
		WorkspaceID: "11111111-1111-1111-1111-111111111111",
		UserID:      "22222222-2222-2222-2222-222222222222",
		Role:        "member",
	})

	require.ErrorIs(t, err, members.ErrAlreadyMember)
}

func TestAddMember_StoreError(t *testing.T) {
	expectedErr := errors.New("store error")
	store := &mockMemberStore{
		getMemberFn: func(_ context.Context, _ db.GetMemberParams) (db.WorkspaceMember, error) {
			return db.WorkspaceMember{}, pgx.ErrNoRows
		},
		addMemberFn: func(_ context.Context, _ db.AddMemberParams) (db.WorkspaceMember, error) {
			return db.WorkspaceMember{}, expectedErr
		},
	}

	service := members.NewMemberService(store)
	_, err := service.AddMember(context.Background(), members.AddMemberInput{
		WorkspaceID: "11111111-1111-1111-1111-111111111111",
		UserID:      "22222222-2222-2222-2222-222222222222",
		Role:        "member",
	})

	require.ErrorIs(t, err, expectedErr)
}

func TestListMembers_InvalidWorkspaceID(t *testing.T) {
	store := &mockMemberStore{
		listMembersFn: func(_ context.Context, _ pgtype.UUID) ([]db.WorkspaceMember, error) {
			t.Fatal("ListMembers should not be called with an invalid workspace ID")
			return nil, nil
		},
	}

	service := members.NewMemberService(store)
	_, err := service.ListMembers(context.Background(), "not-a-uuid")

	require.ErrorIs(t, err, common.ErrInvalidWorkspaceID)
}

func TestListMembers_StoreError(t *testing.T) {
	expectedErr := errors.New("store error")
	store := &mockMemberStore{
		listMembersFn: func(_ context.Context, workspaceID pgtype.UUID) ([]db.WorkspaceMember, error) {
			require.Equal(t, "11111111-1111-1111-1111-111111111111", workspaceID.String())
			return nil, expectedErr
		},
	}

	service := members.NewMemberService(store)
	_, err := service.ListMembers(context.Background(), "11111111-1111-1111-1111-111111111111")

	require.ErrorIs(t, err, expectedErr)
}

func TestListMembers_Success(t *testing.T) {
	expectedMembers := []db.WorkspaceMember{
		{
			WorkspaceID: mustUUID(t, "11111111-1111-1111-1111-111111111111"),
			UserID:      mustUUID(t, "22222222-2222-2222-2222-222222222222"),
			Role:        "administrator",
		},
		{
			WorkspaceID: mustUUID(t, "11111111-1111-1111-1111-111111111111"),
			UserID:      mustUUID(t, "33333333-3333-3333-3333-333333333333"),
			Role:        "member",
		},
	}
	store := &mockMemberStore{
		listMembersFn: func(_ context.Context, workspaceID pgtype.UUID) ([]db.WorkspaceMember, error) {
			require.Equal(t, "11111111-1111-1111-1111-111111111111", workspaceID.String())
			return expectedMembers, nil
		},
	}

	service := members.NewMemberService(store)
	memberships, err := service.ListMembers(context.Background(), "11111111-1111-1111-1111-111111111111")

	require.NoError(t, err)
	require.Equal(t, expectedMembers, memberships)
}
