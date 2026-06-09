package workspaces_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ottermq/otterboard/src/backend/internal/common"
	"github.com/ottermq/otterboard/src/backend/internal/db"
	"github.com/ottermq/otterboard/src/backend/internal/workspaces"
	"github.com/stretchr/testify/require"
)

type mockWorkspaceStore struct {
	createWorkspaceFn         func(ctx context.Context, arg db.CreateWorkspaceParams) (db.Workspace, error)
	getWorkspaceByIDFn        func(ctx context.Context, id pgtype.UUID) (db.Workspace, error)
	getWorkspacesByMemberIDFn func(ctx context.Context, memberID pgtype.UUID) ([]db.Workspace, error)
	updateWorkspaceFn         func(ctx context.Context, arg db.UpdateWorkspaceParams) (db.Workspace, error)
	deleteWorkspaceFn         func(ctx context.Context, id pgtype.UUID) error
	addMemberFn               func(ctx context.Context, arg db.AddMemberParams) (db.WorkspaceMember, error)
}

func (m *mockWorkspaceStore) CreateWorkspace(ctx context.Context, arg db.CreateWorkspaceParams) (db.Workspace, error) {
	return m.createWorkspaceFn(ctx, arg)
}

func (m *mockWorkspaceStore) GetWorkspaceByID(ctx context.Context, id pgtype.UUID) (db.Workspace, error) {
	return m.getWorkspaceByIDFn(ctx, id)
}

func (m *mockWorkspaceStore) GetWorkspacesByMemberID(ctx context.Context, memberID pgtype.UUID) ([]db.Workspace, error) {
	return m.getWorkspacesByMemberIDFn(ctx, memberID)
}

func (m *mockWorkspaceStore) UpdateWorkspace(ctx context.Context, arg db.UpdateWorkspaceParams) (db.Workspace, error) {
	return m.updateWorkspaceFn(ctx, arg)
}

func (m *mockWorkspaceStore) DeleteWorkspace(ctx context.Context, id pgtype.UUID) error {
	return m.deleteWorkspaceFn(ctx, id)
}

func (m *mockWorkspaceStore) AddMember(ctx context.Context, arg db.AddMemberParams) (db.WorkspaceMember, error) {
	return m.addMemberFn(ctx, arg)
}

func mustUUID(t *testing.T, id string) pgtype.UUID {
	t.Helper()

	var uuid pgtype.UUID
	require.NoError(t, uuid.Scan(id))
	return uuid
}

func TestCreateWorkspace_Success(t *testing.T) {
	ownerID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	workspaceID := mustUUID(t, "22222222-2222-2222-2222-222222222222")

	store := &mockWorkspaceStore{
		createWorkspaceFn: func(_ context.Context, arg db.CreateWorkspaceParams) (db.Workspace, error) {
			return db.Workspace{
				ID:      workspaceID,
				Name:    arg.Name,
				OwnerID: arg.OwnerID,
			}, nil
		},
		addMemberFn: func(_ context.Context, arg db.AddMemberParams) (db.WorkspaceMember, error) {
			require.Equal(t, workspaceID, arg.WorkspaceID)
			require.Equal(t, ownerID, arg.UserID)
			require.Equal(t, "administrator", arg.Role)
			return db.WorkspaceMember{}, nil
		},
	}

	service := workspaces.NewWorkspaceService(store)
	workspace, err := service.CreateWorkspace(context.Background(), workspaces.CreateWorkspaceInput{
		Name:    "Test Workspace",
		OwnerID: ownerID.String(),
	})
	require.NoError(t, err)
	require.Equal(t, "Test Workspace", workspace.Name)
	require.Equal(t, ownerID.String(), workspace.OwnerID)
}

func TestGetWorkspaceByID_Success(t *testing.T) {
	workspaceID := mustUUID(t, "11111111-1111-1111-1111-111111111111")

	store := &mockWorkspaceStore{
		getWorkspaceByIDFn: func(_ context.Context, id pgtype.UUID) (db.Workspace, error) {
			return db.Workspace{
				ID:   id,
				Name: "Test Workspace",
			}, nil
		},
	}

	service := workspaces.NewWorkspaceService(store)
	ws, err := service.GetWorkspaceByID(context.Background(), workspaces.GetWorkspaceByIdInput{
		ID: workspaceID.String(),
	})
	require.NoError(t, err)
	require.Equal(t, workspaceID.String(), ws.ID)
	require.Equal(t, "Test Workspace", ws.Name)
}

func TestGetWorkspaceByID_NotFound(t *testing.T) {
	workspaceID := mustUUID(t, "11111111-1111-1111-1111-111111111111")

	store := &mockWorkspaceStore{
		getWorkspaceByIDFn: func(_ context.Context, id pgtype.UUID) (db.Workspace, error) {
			return db.Workspace{}, pgx.ErrNoRows
		},
	}

	service := workspaces.NewWorkspaceService(store)
	_, err := service.GetWorkspaceByID(context.Background(), workspaces.GetWorkspaceByIdInput{
		ID: workspaceID.String(),
	})
	require.ErrorIs(t, err, workspaces.ErrWorkspaceNotFound)
}

func TestCreateWorkspace_InvalidOwnerID(t *testing.T) {
	store := &mockWorkspaceStore{
		createWorkspaceFn: func(_ context.Context, arg db.CreateWorkspaceParams) (db.Workspace, error) {
			return db.Workspace{
				Name:    arg.Name,
				OwnerID: pgtype.UUID{},
			}, nil
		},
	}

	service := workspaces.NewWorkspaceService(store)
	_, err := service.CreateWorkspace(context.Background(), workspaces.CreateWorkspaceInput{
		Name:    "Test Workspace",
		OwnerID: "invalid-uuid",
	})
	require.Error(t, err)
	require.ErrorIs(t, err, workspaces.ErrInvalidOwnerID)
}

func TestCreateWorkspace_AddMemberError(t *testing.T) {
	ownerID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	expectedErr := errors.New("add member failed")

	store := &mockWorkspaceStore{
		createWorkspaceFn: func(_ context.Context, arg db.CreateWorkspaceParams) (db.Workspace, error) {
			return db.Workspace{
				ID:      mustUUID(t, "22222222-2222-2222-2222-222222222222"),
				Name:    arg.Name,
				OwnerID: arg.OwnerID,
			}, nil
		},
		addMemberFn: func(_ context.Context, _ db.AddMemberParams) (db.WorkspaceMember, error) {
			return db.WorkspaceMember{}, expectedErr
		},
	}

	service := workspaces.NewWorkspaceService(store)
	_, err := service.CreateWorkspace(context.Background(), workspaces.CreateWorkspaceInput{
		Name:    "Test Workspace",
		OwnerID: ownerID.String(),
	})

	require.ErrorIs(t, err, expectedErr)
}

func TestGetWorkspacesByMemberID_Success(t *testing.T) {
	memberID := mustUUID(t, "123e4567-e89b-12d3-a456-426614174000")

	store := &mockWorkspaceStore{
		getWorkspacesByMemberIDFn: func(_ context.Context, id pgtype.UUID) ([]db.Workspace, error) {
			require.Equal(t, memberID, id)
			return []db.Workspace{
				{
					ID:      mustUUID(t, "89e41058-2e01-487e-ad8c-8e9e35c31987"),
					Name:    "Workspace 1",
					OwnerID: mustUUID(t, "11111111-1111-1111-1111-111111111111"),
				},
				{
					ID:      mustUUID(t, "98e41058-2e01-487e-ad8c-8e9e35c31988"),
					Name:    "Workspace 2",
					OwnerID: mustUUID(t, "22222222-2222-2222-2222-222222222222"),
				},
			}, nil
		},
	}

	service := workspaces.NewWorkspaceService(store)
	result, err := service.GetWorkspacesByMemberID(context.Background(), memberID.String())

	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Equal(t, "Workspace 1", result[0].Name)
	require.Equal(t, "Workspace 2", result[1].Name)
}

func TestGetWorkspacesByMemberID_InvalidMemberID(t *testing.T) {
	service := workspaces.NewWorkspaceService(&mockWorkspaceStore{})

	_, err := service.GetWorkspacesByMemberID(context.Background(), "invalid-uuid")

	require.ErrorIs(t, err, workspaces.ErrInvalidMemberID)
}

func TestGetWorkspacesByMemberID_StoreError(t *testing.T) {
	expectedErr := errors.New("list member workspaces failed")

	store := &mockWorkspaceStore{
		getWorkspacesByMemberIDFn: func(_ context.Context, _ pgtype.UUID) ([]db.Workspace, error) {
			return nil, expectedErr
		},
	}

	service := workspaces.NewWorkspaceService(store)
	_, err := service.GetWorkspacesByMemberID(context.Background(), "123e4567-e89b-12d3-a456-426614174000")

	require.ErrorIs(t, err, expectedErr)
}

func TestUpdateWorkspace_Success(t *testing.T) {
	workspaceID := mustUUID(t, "89e41058-2e01-487e-ad8c-8e9e35c31987")

	store := &mockWorkspaceStore{
		getWorkspaceByIDFn: func(_ context.Context, id pgtype.UUID) (db.Workspace, error) {
			return db.Workspace{
				ID:   id,
				Name: "Test Workspace",
			}, nil
		},
		updateWorkspaceFn: func(_ context.Context, arg db.UpdateWorkspaceParams) (db.Workspace, error) {
			return db.Workspace{
				ID:   arg.ID,
				Name: arg.Name,
			}, nil
		},
	}

	service := workspaces.NewWorkspaceService(store)
	updatedWorkspace, err := service.UpdateWorkspace(context.Background(), workspaces.UpdateWorkspaceInput{
		ID:   workspaceID.String(),
		Name: "Updated Workspace Name",
	})
	require.NoError(t, err)
	require.Equal(t, workspaceID.String(), updatedWorkspace.ID)
	require.Equal(t, "Updated Workspace Name", updatedWorkspace.Name)
}

func TestUpdateWorkspace_InvalidWorkspaceID(t *testing.T) {
	store := &mockWorkspaceStore{
		updateWorkspaceFn: func(_ context.Context, arg db.UpdateWorkspaceParams) (db.Workspace, error) {
			return db.Workspace{}, nil
		},
	}

	service := workspaces.NewWorkspaceService(store)
	_, err := service.UpdateWorkspace(context.Background(), workspaces.UpdateWorkspaceInput{
		ID:   "invalid-uuid",
		Name: "Updated Workspace Name",
	})
	require.Error(t, err)
	require.ErrorIs(t, err, common.ErrInvalidWorkspaceID)
}

func TestUpdateWorkspace_WorkspaceNotFound(t *testing.T) {
	workspaceID := mustUUID(t, "89e41058-2e01-487e-ad8c-8e9e35c31987")
	store := &mockWorkspaceStore{
		updateWorkspaceFn: func(_ context.Context, arg db.UpdateWorkspaceParams) (db.Workspace, error) {
			return db.Workspace{}, pgx.ErrNoRows
		},
	}

	service := workspaces.NewWorkspaceService(store)
	_, err := service.UpdateWorkspace(context.Background(), workspaces.UpdateWorkspaceInput{
		ID:   workspaceID.String(),
		Name: "Updated Workspace Name",
	})
	require.ErrorIs(t, err, workspaces.ErrWorkspaceNotFound)
}

func TestDeleteWorkspace_Success(t *testing.T) {
	workspaceID := mustUUID(t, "89e41058-2e01-487e-ad8c-8e9e35c31987")

	deleteCalled := false
	store := &mockWorkspaceStore{
		getWorkspaceByIDFn: func(_ context.Context, id pgtype.UUID) (db.Workspace, error) {
			require.Equal(t, workspaceID, id)
			return db.Workspace{
				ID:   id,
				Name: "Test Workspace",
			}, nil
		},
		deleteWorkspaceFn: func(_ context.Context, id pgtype.UUID) error {
			require.Equal(t, workspaceID, id)
			deleteCalled = true
			return nil
		},
	}

	service := workspaces.NewWorkspaceService(store)
	err := service.DeleteWorkspace(context.Background(), workspaces.DeleteWorkspaceInput{
		ID: workspaceID.String(),
	})
	require.NoError(t, err)
	require.True(t, deleteCalled)
}

func TestDeleteWorkspace_InvalidWorkspaceID(t *testing.T) {
	store := &mockWorkspaceStore{}

	service := workspaces.NewWorkspaceService(store)
	err := service.DeleteWorkspace(context.Background(), workspaces.DeleteWorkspaceInput{
		ID: "invalid-uuid",
	})
	require.Error(t, err)
	require.ErrorIs(t, err, common.ErrInvalidWorkspaceID)
}

func TestDeleteWorkspace_StoreError(t *testing.T) {
	workspaceID := mustUUID(t, "89e41058-2e01-487e-ad8c-8e9e35c31987")

	expectedErr := errors.New("delete workspace failed")

	store := &mockWorkspaceStore{
		getWorkspaceByIDFn: func(_ context.Context, id pgtype.UUID) (db.Workspace, error) {
			return db.Workspace{
				ID:   id,
				Name: "Test Workspace",
			}, nil
		},
		deleteWorkspaceFn: func(_ context.Context, id pgtype.UUID) error {
			require.Equal(t, workspaceID, id)
			return expectedErr
		},
	}

	service := workspaces.NewWorkspaceService(store)
	err := service.DeleteWorkspace(context.Background(), workspaces.DeleteWorkspaceInput{
		ID: workspaceID.String(),
	})
	require.ErrorIs(t, err, expectedErr)
}
