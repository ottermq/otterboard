package projects_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ottermq/otterboard/src/backend/internal/common"
	"github.com/ottermq/otterboard/src/backend/internal/db"
	"github.com/ottermq/otterboard/src/backend/internal/projects"
	"github.com/stretchr/testify/require"
)

type mockProjectStore struct {
	createProjectFn func(ctx context.Context, arg db.CreateProjectParams) (db.Project, error)
}

func (m *mockProjectStore) CreateProject(ctx context.Context, arg db.CreateProjectParams) (db.Project, error) {
	return m.createProjectFn(ctx, arg)
}

func mustUUID(t *testing.T, id string) pgtype.UUID {
	t.Helper()

	var uuid pgtype.UUID
	require.NoError(t, uuid.Scan(id))
	return uuid
}

func TestCreateProject_Success(t *testing.T) {
	workspaceID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	projectID := mustUUID(t, "22222222-2222-2222-2222-222222222222")

	store := &mockProjectStore{
		createProjectFn: func(_ context.Context, arg db.CreateProjectParams) (db.Project, error) {
			return db.Project{
				ID:          projectID,
				WorkspaceID: arg.WorkspaceID,
				Name:        arg.Name,
				ImageUrl:    arg.ImageUrl,
				CreatedAt:   pgtype.Timestamptz{},
				UpdatedAt:   pgtype.Timestamptz{},
			}, nil
		},
	}

	service := projects.NewProjectService(store)
	project, err := service.CreateProject(context.Background(), projects.CreateProjectInput{
		WorkspaceID: workspaceID.String(),
		Name:        "Test Project",
		ImageUrl:    "test_image_url",
	})
	require.NoError(t, err)
	require.Equal(t, workspaceID.String(), project.WorkspaceID)
	require.Equal(t, "Test Project", project.Name)
}

func TestCreateProject_InvalidWorkspaceID(t *testing.T) {
	store := &mockProjectStore{
		createProjectFn: func(_ context.Context, arg db.CreateProjectParams) (db.Project, error) {
			t.Fatal("CreateProject should not be called with invalid input")
			return db.Project{}, nil
		},
	}

	service := projects.NewProjectService(store)
	project, err := service.CreateProject(context.Background(), projects.CreateProjectInput{
		WorkspaceID: "invalid UUID",
		Name:        "Invalid UUID Project",
	})
	require.ErrorIs(t, err, common.ErrInvalidWorkspaceID)
	require.Equal(t, project, projects.Project{})
}

func TestCreateProject_EmptyName(t *testing.T) {
	workspaceID := mustUUID(t, "11111111-1111-1111-1111-111111111111")

	store := &mockProjectStore{
		createProjectFn: func(_ context.Context, arg db.CreateProjectParams) (db.Project, error) {
			t.Fatal("CreateProject should not be called with invalid input")
			return db.Project{}, nil
		},
	}

	service := projects.NewProjectService(store)
	project, err := service.CreateProject(context.Background(), projects.CreateProjectInput{
		WorkspaceID: workspaceID.String(),
		Name:        "",
	})
	require.ErrorIs(t, err, projects.ErrInvalidProjectName)
	require.Equal(t, project, projects.Project{})
}

func TestCreateProject_StoreError(t *testing.T) {
	workspaceID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	expectedErr := errors.New("generic storage error")

	store := &mockProjectStore{
		createProjectFn: func(_ context.Context, arg db.CreateProjectParams) (db.Project, error) {
			return db.Project{}, expectedErr
		},
	}

	service := projects.NewProjectService(store)
	project, err := service.CreateProject(context.Background(), projects.CreateProjectInput{
		WorkspaceID: workspaceID.String(),
		Name:        "Generic Name",
	})
	require.ErrorIs(t, err, expectedErr)
	require.Equal(t, project, projects.Project{})
}
