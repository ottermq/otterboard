package issues_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ottermq/otterboard/src/backend/internal/common"
	"github.com/ottermq/otterboard/src/backend/internal/db"
	"github.com/ottermq/otterboard/src/backend/internal/issues"
	"github.com/stretchr/testify/require"
)

type mockIssueStore struct {
	getMaxPositionByProjectAndStatusFn func(ctx context.Context, arg db.GetMaxPositionByProjectAndStatusParams) (any, error)
	createIssueFn                      func(ctx context.Context, arg db.CreateIssueParams) (db.Issue, error)
	getIssueByIDFn                     func(ctx context.Context, arg db.GetIssueByIDParams) (db.Issue, error)
	listIssuesByProjectFn              func(ctx context.Context, arg db.ListIssuesByProjectParams) ([]db.Issue, error)
	countIssuesByProjectFn             func(ctx context.Context, projectID pgtype.UUID) (int64, error)
	listIssuesByWorkspaceFn            func(ctx context.Context, arg db.ListIssuesByWorkspaceParams) ([]db.Issue, error)
	countIssuesByWorkspaceFn           func(ctx context.Context, workspaceID pgtype.UUID) (int64, error)
	updateIssueFn                      func(ctx context.Context, arg db.UpdateIssueParams) (db.Issue, error)
	deleteIssueFn                      func(ctx context.Context, arg db.DeleteIssueParams) error
}

func (m *mockIssueStore) GetMaxPositionByProjectAndStatus(ctx context.Context, arg db.GetMaxPositionByProjectAndStatusParams) (any, error) {
	if m.getMaxPositionByProjectAndStatusFn == nil {
		panic("unexpected call to GetMaxPositionByProjectAndStatus")
	}
	return m.getMaxPositionByProjectAndStatusFn(ctx, arg)
}

func (m *mockIssueStore) CreateIssue(ctx context.Context, arg db.CreateIssueParams) (db.Issue, error) {
	if m.createIssueFn == nil {
		panic("unexpected call to CreateIssue")
	}
	return m.createIssueFn(ctx, arg)
}

func (m *mockIssueStore) GetIssueByID(ctx context.Context, arg db.GetIssueByIDParams) (db.Issue, error) {
	if m.getIssueByIDFn == nil {
		panic("unexpected call to GetIssueByID")
	}
	return m.getIssueByIDFn(ctx, arg)
}

func (m *mockIssueStore) ListIssuesByProject(ctx context.Context, arg db.ListIssuesByProjectParams) ([]db.Issue, error) {
	if m.listIssuesByProjectFn == nil {
		panic("unexpected call to ListIssuesByProject")
	}
	return m.listIssuesByProjectFn(ctx, arg)
}

func (m *mockIssueStore) CountIssuesByProject(ctx context.Context, projectID pgtype.UUID) (int64, error) {
	if m.countIssuesByProjectFn == nil {
		panic("unexpected call to CountIssuesByProject")
	}
	return m.countIssuesByProjectFn(ctx, projectID)
}

func (m *mockIssueStore) ListIssuesByWorkspace(ctx context.Context, arg db.ListIssuesByWorkspaceParams) ([]db.Issue, error) {
	if m.listIssuesByWorkspaceFn == nil {
		panic("unexpected call to ListIssuesByWorkspace")
	}
	return m.listIssuesByWorkspaceFn(ctx, arg)
}

func (m *mockIssueStore) CountIssuesByWorkspace(ctx context.Context, workspaceID pgtype.UUID) (int64, error) {
	if m.countIssuesByWorkspaceFn == nil {
		panic("unexpected call to CountIssuesByWorkspace")
	}
	return m.countIssuesByWorkspaceFn(ctx, workspaceID)
}

func (m *mockIssueStore) UpdateIssue(ctx context.Context, arg db.UpdateIssueParams) (db.Issue, error) {
	if m.updateIssueFn == nil {
		panic("unexpected call to UpdateIssue")
	}
	return m.updateIssueFn(ctx, arg)
}

func (m *mockIssueStore) DeleteIssue(ctx context.Context, arg db.DeleteIssueParams) error {
	if m.deleteIssueFn == nil {
		panic("unexpected call to DeleteIssue")
	}
	return m.deleteIssueFn(ctx, arg)
}

func mustUUID(t *testing.T, id string) pgtype.UUID {
	t.Helper()

	var uuid pgtype.UUID
	require.NoError(t, uuid.Scan(id))
	return uuid
}

func TestCreateIssue_Success(t *testing.T) {
	projectID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	issueID := mustUUID(t, "22222222-2222-2222-2222-222222222222")
	userID := mustUUID(t, "33333333-3333-3333-3333-333333333333")
	store := &mockIssueStore{
		getMaxPositionByProjectAndStatusFn: func(ctx context.Context, arg db.GetMaxPositionByProjectAndStatusParams) (any, error) {
			return float64(0.0), nil
		},
		createIssueFn: func(ctx context.Context, arg db.CreateIssueParams) (db.Issue, error) {
			return db.Issue{
				ID:         issueID,
				ProjectID:  arg.ProjectID,
				Title:      arg.Title,
				Overview:   arg.Overview,
				Type:       arg.Type,
				Status:     arg.Status,
				Position:   arg.Position,
				AssigneeID: arg.AssigneeID,
				CreatedBy:  arg.CreatedBy,
				DueDate:    arg.DueDate,
				CreatedAt:  pgtype.Timestamptz{},
				UpdatedAt:  pgtype.Timestamptz{},
			}, nil
		},
	}
	service := issues.NewIssueService(store)
	issue, err := service.CreateIssue(context.Background(), issues.CreateIssueInput{
		ProjectID:  projectID.String(),
		Title:      "Test Issue",
		Overview:   "This is a test",
		Type:       "bug",
		AssigneeID: userID.String(),
		CreatedBy:  userID.String(),
		DueDate:    nil,
	})
	require.NoError(t, err)
	require.Equal(t, issueID.String(), issue.ID)
}

func TestCreateIssue_ValidationErrors(t *testing.T) {
	validProjectID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	validUserID := mustUUID(t, "22222222-2222-2222-2222-222222222222")
	tests := []struct {
		name      string
		input     issues.CreateIssueInput
		wantError error
	}{
		{
			name: "invalid project ID",
			input: issues.CreateIssueInput{
				ProjectID:  "invalid UUID",
				Title:      "title",
				Type:       "bug",
				AssigneeID: validUserID.String(),
				CreatedBy:  validUserID.String(),
				DueDate:    &time.Time{},
			},
			wantError: common.ErrInvalidProjectID,
		},
		{
			name: "invalid assignee Id",
			input: issues.CreateIssueInput{
				ProjectID:  validProjectID.String(),
				Title:      "title",
				Type:       "bug",
				AssigneeID: "invalid assignee",
				CreatedBy:  validUserID.String(),
				DueDate:    &time.Time{},
			},
			wantError: issues.ErrInvalidAssigneeID,
		},
		{
			name: "invalid user Id",
			input: issues.CreateIssueInput{
				ProjectID:  validProjectID.String(),
				Title:      "title",
				Type:       "bug",
				AssigneeID: validUserID.String(),
				CreatedBy:  "invalid user ID",
				DueDate:    &time.Time{},
			},
			wantError: common.ErrInvalidUserID,
		},
		{
			name: "invalid title",
			input: issues.CreateIssueInput{
				ProjectID:  validProjectID.String(),
				Title:      "",
				Type:       "bug",
				AssigneeID: validUserID.String(),
				CreatedBy:  validUserID.String(),
				DueDate:    &time.Time{},
			},
			wantError: issues.ErrInvalidTitle,
		},
		{
			name: "invalid issue type",
			input: issues.CreateIssueInput{
				ProjectID:  validProjectID.String(),
				Title:      "title",
				Type:       "",
				AssigneeID: validUserID.String(),
				CreatedBy:  validUserID.String(),
				DueDate:    &time.Time{},
			},
			wantError: issues.ErrInvalidType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockIssueStore{
				createIssueFn: func(ctx context.Context, arg db.CreateIssueParams) (db.Issue, error) {
					t.Fatal("CreateIssue should not be called with invalid input")
					return db.Issue{}, nil
				},
			}

			service := issues.NewIssueService(store)
			got, err := service.CreateIssue(context.Background(), tt.input)
			require.ErrorIs(t, err, tt.wantError)
			require.Equal(t, issues.Issue{}, got)
		})
	}
}

func TestCreateIssue_StoreError(t *testing.T) {
	projectID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	userID := mustUUID(t, "33333333-3333-3333-3333-333333333333")
	expectedErr := errors.New("generic storage error")
	store := &mockIssueStore{
		getMaxPositionByProjectAndStatusFn: func(ctx context.Context, arg db.GetMaxPositionByProjectAndStatusParams) (any, error) {
			return float64(0.0), nil
		},
		createIssueFn: func(ctx context.Context, arg db.CreateIssueParams) (db.Issue, error) {
			return db.Issue{}, expectedErr
		},
	}
	service := issues.NewIssueService(store)
	issue, err := service.CreateIssue(context.Background(), issues.CreateIssueInput{
		ProjectID:  projectID.String(),
		Title:      "Test Issue",
		Overview:   "This is a test",
		Type:       "bug",
		AssigneeID: userID.String(),
		CreatedBy:  userID.String(),
		DueDate:    nil,
	})
	require.ErrorIs(t, err, expectedErr)
	require.Equal(t, issues.Issue{}, issue)
}

func TestGetIssueByID_Success(t *testing.T) {
	projectID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	issueID := mustUUID(t, "22222222-2222-2222-2222-222222222222")

	store := &mockIssueStore{
		getIssueByIDFn: func(ctx context.Context, arg db.GetIssueByIDParams) (db.Issue, error) {
			return db.Issue{
				ID:        issueID,
				ProjectID: projectID,
			}, nil
		},
	}
	service := issues.NewIssueService(store)
	got, err := service.GetIssueByID(context.Background(), issues.GetIssueByIdInput{
		ID:        issueID.String(),
		ProjectID: projectID.String(),
	})
	require.NoError(t, err)
	require.Equal(t, issueID.String(), got.ID)
	require.Equal(t, projectID.String(), got.ProjectID)
}

func TestGetIssueByID_ValidationErrors(t *testing.T) {
	validProjectID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	validIssueID := mustUUID(t, "22222222-2222-2222-2222-222222222222")
	tests := []struct {
		name    string
		input   issues.GetIssueByIdInput
		wantErr error
	}{
		{
			name: "invalid issue ID",
			input: issues.GetIssueByIdInput{
				ID:        "invalid UUID",
				ProjectID: validProjectID.String(),
			},
			wantErr: issues.ErrInvalidIssueID,
		},
		{
			name: "invalid project ID",
			input: issues.GetIssueByIdInput{
				ID:        validIssueID.String(),
				ProjectID: "invalid UUID",
			},
			wantErr: common.ErrInvalidProjectID,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockIssueStore{
				getIssueByIDFn: func(ctx context.Context, arg db.GetIssueByIDParams) (db.Issue, error) {
					t.Fatal("GetIssueById should not be called with invalid params")
					return db.Issue{}, nil
				},
			}

			service := issues.NewIssueService(store)
			got, err := service.GetIssueByID(context.Background(), tt.input)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, issues.Issue{}, got)
		})
	}
}

func TestGetIssueByID_NotFound(t *testing.T) {
	issueID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	projectID := mustUUID(t, "22222222-2222-2222-2222-222222222222")

	store := &mockIssueStore{
		getIssueByIDFn: func(ctx context.Context, arg db.GetIssueByIDParams) (db.Issue, error) {
			return db.Issue{}, pgx.ErrNoRows
		},
	}
	service := issues.NewIssueService(store)
	got, err := service.GetIssueByID(context.Background(), issues.GetIssueByIdInput{
		ID:        issueID.String(),
		ProjectID: projectID.String(),
	})
	require.ErrorIs(t, err, issues.ErrIssueNotFound)
	require.Equal(t, issues.Issue{}, got)
}

func TestGetIssueByID_StoreError(t *testing.T) {
	issueID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	projectID := mustUUID(t, "22222222-2222-2222-2222-222222222222")
	expectedErr := errors.New("generic error")

	store := &mockIssueStore{
		getIssueByIDFn: func(ctx context.Context, arg db.GetIssueByIDParams) (db.Issue, error) {
			return db.Issue{}, expectedErr
		},
	}
	service := issues.NewIssueService(store)
	got, err := service.GetIssueByID(context.Background(), issues.GetIssueByIdInput{
		ID:        issueID.String(),
		ProjectID: projectID.String(),
	})
	require.ErrorIs(t, err, expectedErr)
	require.Equal(t, issues.Issue{}, got)
}

func TestListIssuesByProject_Success(t *testing.T) {
	projectID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	issueID := mustUUID(t, "22222222-2222-2222-2222-222222222222")

	store := &mockIssueStore{
		listIssuesByProjectFn: func(ctx context.Context, arg db.ListIssuesByProjectParams) ([]db.Issue, error) {
			return []db.Issue{{
				ID:        issueID,
				ProjectID: projectID,
			}}, nil
		},
	}
	service := issues.NewIssueService(store)
	got, err := service.ListIssuesByProject(context.Background(), issues.ListIssuesByProjectInput{
		ProjectID: projectID.String(),
		Page:      1,
		Limit:     1,
	})
	require.NoError(t, err)
	require.Equal(t, issueID.String(), got[0].ID)
}

func TestListIssuesByProject_InvalidProjectID(t *testing.T) {
	store := &mockIssueStore{
		listIssuesByProjectFn: func(ctx context.Context, arg db.ListIssuesByProjectParams) ([]db.Issue, error) {
			t.Fatal("ListIssuesByProject should not be called with invalid input")
			return []db.Issue{}, nil
		},
	}
	service := issues.NewIssueService(store)
	got, err := service.ListIssuesByProject(context.Background(), issues.ListIssuesByProjectInput{
		ProjectID: "invalid UUID",
		Page:      1,
		Limit:     1,
	})
	require.ErrorIs(t, err, common.ErrInvalidProjectID)
	require.Equal(t, []issues.Issue{}, got)
}

func TestListIssuesByProject_StoreError(t *testing.T) {
	expectErr := errors.New("generic error")
	store := &mockIssueStore{
		listIssuesByProjectFn: func(ctx context.Context, arg db.ListIssuesByProjectParams) ([]db.Issue, error) {
			return []db.Issue{}, expectErr
		},
	}
	service := issues.NewIssueService(store)
	got, err := service.ListIssuesByProject(context.Background(), issues.ListIssuesByProjectInput{
		ProjectID: "11111111-1111-1111-1111-111111111111",
		Page:      1,
		Limit:     1,
	})
	require.ErrorIs(t, err, expectErr)
	require.Equal(t, []issues.Issue{}, got)
}

func TestListIssuesByProject_Pagination(t *testing.T) {
	projectID := mustUUID(t, "11111111-1111-1111-1111-111111111111")

	tests := []struct {
		name       string
		page       int32
		limit      int32
		wantLimit  int32
		wantOffset int32
	}{
		{
			name:       "first page",
			page:       1,
			limit:      10,
			wantLimit:  10,
			wantOffset: 0,
		},
		{
			name:       "second page",
			page:       2,
			limit:      10,
			wantLimit:  10,
			wantOffset: 10,
		},
		{
			name:       "third page with different limit",
			page:       3,
			limit:      25,
			wantLimit:  25,
			wantOffset: 50,
		},
		{
			name:       "forth page with invalid limit",
			page:       4,
			limit:      -1,
			wantLimit:  issues.DefaultLimit,
			wantOffset: 60,
		},
		{
			name:       "no page with invalid limit",
			limit:      -1,
			wantLimit:  issues.DefaultLimit,
			wantOffset: 0,
		},
		{
			name:       "limit beyond max",
			limit:      200,
			wantLimit:  issues.MaxLimit,
			wantOffset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockIssueStore{
				listIssuesByProjectFn: func(ctx context.Context, arg db.ListIssuesByProjectParams) ([]db.Issue, error) {
					require.Equal(t, projectID, arg.ProjectID)
					require.Equal(t, tt.wantLimit, arg.Limit)
					require.Equal(t, tt.wantOffset, arg.Offset)
					return []db.Issue{}, nil
				},
			}

			service := issues.NewIssueService(store)
			got, err := service.ListIssuesByProject(context.Background(), issues.ListIssuesByProjectInput{
				ProjectID: projectID.String(),
				Page:      tt.page,
				Limit:     tt.limit,
			})
			require.NoError(t, err)
			require.Empty(t, got)
		})
	}
}

func TestCountIssuesByProject_Success(t *testing.T) {
	projectID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	store := &mockIssueStore{
		countIssuesByProjectFn: func(ctx context.Context, projectID pgtype.UUID) (int64, error) {
			return 1, nil
		},
	}

	service := issues.NewIssueService(store)
	count, err := service.CountIssuesByProject(context.Background(), projectID.String())
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
}

func TestCountIssuesByProject_InvalidProjectID(t *testing.T) {
	store := &mockIssueStore{
		countIssuesByProjectFn: func(ctx context.Context, projectID pgtype.UUID) (int64, error) {
			t.Fatal("CountIssuesByProject should not be called with invalid input")
			return 0, nil
		},
	}

	service := issues.NewIssueService(store)
	count, err := service.CountIssuesByProject(context.Background(), "invalid UUID")
	require.ErrorIs(t, err, common.ErrInvalidProjectID)
	require.Equal(t, int64(0), count)
}

func TestListIssuesByWorkspace_Success(t *testing.T) {
	workspaceID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	projectID := mustUUID(t, "22222222-2222-2222-2222-222222222222")
	issueID := mustUUID(t, "33333333-3333-3333-3333-333333333333")

	store := &mockIssueStore{
		listIssuesByWorkspaceFn: func(ctx context.Context, arg db.ListIssuesByWorkspaceParams) ([]db.Issue, error) {
			return []db.Issue{
				{
					ID:        issueID,
					ProjectID: projectID,
				},
			}, nil
		},
	}
	service := issues.NewIssueService(store)
	got, err := service.ListIssuesByWorkspace(context.Background(), issues.ListIssuesByWorkspaceInput{
		WorkspaceID: workspaceID.String(),
		Page:        1,
		Limit:       1,
	})
	require.NoError(t, err)
	require.Equal(t, issueID.String(), got[0].ID)
}

func TestListIssuesByWorkspace_InvalidWorkspaceID(t *testing.T) {

	store := &mockIssueStore{
		listIssuesByWorkspaceFn: func(ctx context.Context, arg db.ListIssuesByWorkspaceParams) ([]db.Issue, error) {
			t.Fatal("ListIssuesByWorkspace should not be called with invalid input")
			return []db.Issue{}, nil
		},
	}
	service := issues.NewIssueService(store)
	got, err := service.ListIssuesByWorkspace(context.Background(), issues.ListIssuesByWorkspaceInput{
		WorkspaceID: "invalid UUID",
		Page:        1,
		Limit:       1,
	})
	require.ErrorIs(t, err, common.ErrInvalidWorkspaceID)
	require.Equal(t, []issues.Issue{}, got)
}

func TestListIssuesByWorkspace_StoreError(t *testing.T) {
	workspaceID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	expectedErr := errors.New("generic error")
	store := &mockIssueStore{
		listIssuesByWorkspaceFn: func(ctx context.Context, arg db.ListIssuesByWorkspaceParams) ([]db.Issue, error) {
			return []db.Issue{}, expectedErr
		},
	}
	service := issues.NewIssueService(store)
	got, err := service.ListIssuesByWorkspace(context.Background(), issues.ListIssuesByWorkspaceInput{
		WorkspaceID: workspaceID.String(),
		Page:        1,
		Limit:       1,
	})
	require.ErrorIs(t, err, expectedErr)
	require.Equal(t, []issues.Issue{}, got)
}

func TestListIssuesByWorkspace_Pagination(t *testing.T) {
	workspaceID := mustUUID(t, "11111111-1111-1111-1111-111111111111")

	tests := []struct {
		name       string
		page       int32
		limit      int32
		wantLimit  int32
		wantOffset int32
	}{
		{
			name:       "first page",
			page:       1,
			limit:      10,
			wantLimit:  10,
			wantOffset: 0,
		},
		{
			name:       "second page",
			page:       2,
			limit:      10,
			wantLimit:  10,
			wantOffset: 10,
		},
		{
			name:       "third page with different limit",
			page:       3,
			limit:      25,
			wantLimit:  25,
			wantOffset: 50,
		},
		{
			name:       "forth page with invalid limit",
			page:       4,
			limit:      -1,
			wantLimit:  issues.DefaultLimit,
			wantOffset: 60,
		},
		{
			name:       "no page with invalid limit",
			limit:      -1,
			wantLimit:  issues.DefaultLimit,
			wantOffset: 0,
		},
		{
			name:       "limit beyond max",
			limit:      200,
			wantLimit:  issues.MaxLimit,
			wantOffset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockIssueStore{
				listIssuesByWorkspaceFn: func(ctx context.Context, arg db.ListIssuesByWorkspaceParams) ([]db.Issue, error) {
					require.Equal(t, workspaceID, arg.WorkspaceID)
					require.Equal(t, tt.wantLimit, arg.Limit)
					require.Equal(t, tt.wantOffset, arg.Offset)
					return []db.Issue{}, nil
				},
			}

			service := issues.NewIssueService(store)
			got, err := service.ListIssuesByWorkspace(context.Background(), issues.ListIssuesByWorkspaceInput{
				WorkspaceID: workspaceID.String(),
				Page:        tt.page,
				Limit:       tt.limit,
			})
			require.NoError(t, err)
			require.Empty(t, got)
		})
	}
}

func TestCountIssuesByWorkspace_Success(t *testing.T) {
	workspaceUUID := mustUUID(t, "11111111-1111-1111-1111-111111111111")

	store := &mockIssueStore{
		countIssuesByWorkspaceFn: func(ctx context.Context, workspaceID pgtype.UUID) (int64, error) {
			return 1, nil
		},
	}

	service := issues.NewIssueService(store)
	count, err := service.CountIssuesByWorkspace(context.Background(), workspaceUUID.String())
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
}

func TestCountIssuesByWorkspace_InvalidWorkspaceID(t *testing.T) {
	store := &mockIssueStore{
		countIssuesByWorkspaceFn: func(ctx context.Context, workspaceID pgtype.UUID) (int64, error) {
			t.Fatal("CountIssuesByWorkspace should not be called with invalid input")
			return 0, nil
		},
	}

	service := issues.NewIssueService(store)
	count, err := service.CountIssuesByWorkspace(context.Background(), "invalid UUID")
	require.ErrorIs(t, err, common.ErrInvalidWorkspaceID)
	require.Equal(t, int64(0), count)
}

func TestUpdateIssue_Success(t *testing.T) {
	projectID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	issueID := mustUUID(t, "22222222-2222-2222-2222-222222222222")
	userID := mustUUID(t, "33333333-3333-3333-3333-333333333333")
	store := &mockIssueStore{
		updateIssueFn: func(ctx context.Context, arg db.UpdateIssueParams) (db.Issue, error) {
			return db.Issue{
				ID:         issueID,
				ProjectID:  arg.ProjectID,
				Title:      arg.Title,
				Overview:   arg.Overview,
				Type:       arg.Type,
				Status:     arg.Status,
				Position:   arg.Position,
				AssigneeID: arg.AssigneeID,
				CreatedBy:  userID,
				DueDate:    arg.DueDate,
				CreatedAt:  pgtype.Timestamptz{},
				UpdatedAt:  pgtype.Timestamptz{},
			}, nil
		},
	}
	service := issues.NewIssueService(store)
	_, err := service.UpdateIssue(context.Background(), issues.UpdateIssueInput{
		ID:         issueID.String(),
		ProjectID:  projectID.String(),
		Title:      "Test Issue",
		Overview:   "This is a test",
		Type:       "bug",
		Status:     "backlog",
		Position:   float64(1000),
		AssigneeID: userID.String(),
		DueDate:    nil,
	})
	require.NoError(t, err)
}

func TestUpdateIssue_ValidationErrors(t *testing.T) {
	validProjectID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	validUserID := mustUUID(t, "22222222-2222-2222-2222-222222222222")
	validIssueID := mustUUID(t, "33333333-3333-3333-3333-333333333333")
	tests := []struct {
		name      string
		input     issues.UpdateIssueInput
		wantError error
	}{
		{
			name: "invalid issue ID",
			input: issues.UpdateIssueInput{
				ID:         "invalid UUID",
				ProjectID:  validProjectID.String(),
				Title:      "title",
				Type:       "bug",
				Status:     "backlog",
				Position:   float64(1000),
				AssigneeID: validUserID.String(),
				DueDate:    &time.Time{},
			},
			wantError: issues.ErrInvalidIssueID,
		},
		{
			name: "invalid project ID",
			input: issues.UpdateIssueInput{
				ID:         validIssueID.String(),
				ProjectID:  "invalid UUID",
				Title:      "title",
				Type:       "bug",
				Status:     "backlog",
				Position:   float64(1000),
				AssigneeID: validUserID.String(),
				DueDate:    &time.Time{},
			},
			wantError: common.ErrInvalidProjectID,
		},
		{
			name: "invalid assignee Id",
			input: issues.UpdateIssueInput{
				ID:         validIssueID.String(),
				ProjectID:  validProjectID.String(),
				Title:      "title",
				Type:       "bug",
				Status:     "backlog",
				Position:   float64(1000),
				AssigneeID: "invalid UUID",
				DueDate:    &time.Time{},
			},
			wantError: issues.ErrInvalidAssigneeID,
		},
		{
			name: "invalid title",
			input: issues.UpdateIssueInput{
				ID:         validIssueID.String(),
				ProjectID:  validProjectID.String(),
				Title:      "",
				Type:       "bug",
				Status:     "backlog",
				Position:   float64(1000),
				AssigneeID: validUserID.String(),
				DueDate:    &time.Time{},
			},
			wantError: issues.ErrInvalidTitle,
		},
		{
			name: "invalid issue type",
			input: issues.UpdateIssueInput{
				ID:         validIssueID.String(),
				ProjectID:  validProjectID.String(),
				Title:      "title",
				Type:       "invalid",
				Status:     "backlog",
				Position:   float64(1000),
				AssigneeID: validUserID.String(),
				DueDate:    &time.Time{},
			},
			wantError: issues.ErrInvalidType,
		},
		{
			name: "invalid issue status",
			input: issues.UpdateIssueInput{
				ID:         validIssueID.String(),
				ProjectID:  validProjectID.String(),
				Title:      "title",
				Type:       "bug",
				Status:     "invalid",
				Position:   float64(1000),
				AssigneeID: validUserID.String(),
				DueDate:    &time.Time{},
			},
			wantError: issues.ErrInvalidStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockIssueStore{
				updateIssueFn: func(ctx context.Context, arg db.UpdateIssueParams) (db.Issue, error) {
					t.Fatal("UpdateIssue should not be called with invalid input")
					return db.Issue{}, nil
				},
			}

			service := issues.NewIssueService(store)
			got, err := service.UpdateIssue(context.Background(), tt.input)
			require.ErrorIs(t, err, tt.wantError)
			require.Equal(t, issues.Issue{}, got)
		})
	}
}

func TestUpdateIssue_PositionLessThanZero(t *testing.T) {
	projectID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	issueID := mustUUID(t, "22222222-2222-2222-2222-222222222222")
	userID := mustUUID(t, "33333333-3333-3333-3333-333333333333")
	store := &mockIssueStore{
		updateIssueFn: func(ctx context.Context, arg db.UpdateIssueParams) (db.Issue, error) {
			return db.Issue{
				ID:         issueID,
				ProjectID:  arg.ProjectID,
				Title:      arg.Title,
				Overview:   arg.Overview,
				Type:       arg.Type,
				Status:     arg.Status,
				Position:   arg.Position,
				AssigneeID: arg.AssigneeID,
				CreatedBy:  userID,
				DueDate:    arg.DueDate,
				CreatedAt:  pgtype.Timestamptz{},
				UpdatedAt:  pgtype.Timestamptz{},
			}, nil
		},
	}
	service := issues.NewIssueService(store)
	got, err := service.UpdateIssue(context.Background(), issues.UpdateIssueInput{
		ID:         issueID.String(),
		ProjectID:  projectID.String(),
		Title:      "Test Issue",
		Overview:   "This is a test",
		Type:       "bug",
		Status:     "backlog",
		Position:   float64(-1),
		AssigneeID: userID.String(),
		DueDate:    nil,
	})
	require.NoError(t, err)
	require.Equal(t, float64(0), got.Position)
}

func TestUpdateIssue_NotFound(t *testing.T) {
	projectID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	issueID := mustUUID(t, "22222222-2222-2222-2222-222222222222")
	userID := mustUUID(t, "33333333-3333-3333-3333-333333333333")

	store := &mockIssueStore{
		updateIssueFn: func(ctx context.Context, arg db.UpdateIssueParams) (db.Issue, error) {
			return db.Issue{}, pgx.ErrNoRows
		},
	}
	service := issues.NewIssueService(store)
	_, err := service.UpdateIssue(context.Background(), issues.UpdateIssueInput{
		ID:         issueID.String(),
		ProjectID:  projectID.String(),
		Title:      "Test Issue",
		Overview:   "This is a test",
		Type:       "bug",
		Status:     "backlog",
		Position:   float64(1000),
		AssigneeID: userID.String(),
		DueDate:    nil,
	})
	require.ErrorIs(t, err, issues.ErrIssueNotFound)
}

func TestUpdateIssue_StoreError(t *testing.T) {
	projectID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	issueID := mustUUID(t, "22222222-2222-2222-2222-222222222222")
	userID := mustUUID(t, "33333333-3333-3333-3333-333333333333")
	expectedErr := errors.New("generic error")
	store := &mockIssueStore{
		updateIssueFn: func(ctx context.Context, arg db.UpdateIssueParams) (db.Issue, error) {
			return db.Issue{}, expectedErr
		},
	}
	service := issues.NewIssueService(store)
	_, err := service.UpdateIssue(context.Background(), issues.UpdateIssueInput{
		ID:         issueID.String(),
		ProjectID:  projectID.String(),
		Title:      "Test Issue",
		Overview:   "This is a test",
		Type:       "bug",
		Status:     "backlog",
		Position:   float64(1000),
		AssigneeID: userID.String(),
		DueDate:    nil,
	})
	require.ErrorIs(t, err, expectedErr)
}

func TestDeleteIssue_Success(t *testing.T) {
	projectID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	issueID := mustUUID(t, "22222222-2222-2222-2222-222222222222")

	store := &mockIssueStore{
		deleteIssueFn: func(ctx context.Context, arg db.DeleteIssueParams) error {
			require.Equal(t, issueID, arg.ID)
			require.Equal(t, projectID, arg.ProjectID)
			return nil
		},
	}
	service := issues.NewIssueService(store)
	err := service.DeleteIssue(context.Background(), issues.DeleteIssueInput{
		ID:        issueID.String(),
		ProjectID: projectID.String(),
	})
	require.NoError(t, err)
}

func TestDeleteIssue_StoreError(t *testing.T) {
	projectID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	issueID := mustUUID(t, "22222222-2222-2222-2222-222222222222")
	expectedErr := errors.New("generic error")

	store := &mockIssueStore{
		deleteIssueFn: func(ctx context.Context, arg db.DeleteIssueParams) error {
			return expectedErr
		},
	}
	service := issues.NewIssueService(store)
	err := service.DeleteIssue(context.Background(), issues.DeleteIssueInput{
		ID:        issueID.String(),
		ProjectID: projectID.String(),
	})
	require.ErrorIs(t, err, expectedErr)
}
