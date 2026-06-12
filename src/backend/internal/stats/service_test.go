package stats_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ottermq/otterboard/src/backend/internal/common"
	"github.com/ottermq/otterboard/src/backend/internal/db"
	"github.com/ottermq/otterboard/src/backend/internal/stats"
	"github.com/stretchr/testify/require"
)

type mockStatsStore struct {
	getWorkspaceStatsFn func(ctx context.Context, arg db.GetWorkspaceStatsParams) (db.GetWorkspaceStatsRow, error)
}

func (m *mockStatsStore) GetWorkspaceStats(ctx context.Context, arg db.GetWorkspaceStatsParams) (db.GetWorkspaceStatsRow, error) {
	if m.getWorkspaceStatsFn == nil {
		panic("unexpected call to GetWorkspaceStats")
	}
	return m.getWorkspaceStatsFn(ctx, arg)
}

func mustUUID(t *testing.T, id string) pgtype.UUID {
	t.Helper()

	var uuid pgtype.UUID
	require.NoError(t, uuid.Scan(id))
	return uuid
}

func TestGetWorkspaceStats_Success(t *testing.T) {
	workspaceID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	assigneeID := mustUUID(t, "22222222-2222-2222-2222-222222222222")

	store := &mockStatsStore{
		getWorkspaceStatsFn: func(ctx context.Context, arg db.GetWorkspaceStatsParams) (db.GetWorkspaceStatsRow, error) {
			require.Equal(t, workspaceID, arg.WorkspaceID)
			require.Equal(t, assigneeID, arg.AssigneeID)
			return db.GetWorkspaceStatsRow{
				TotalProjects:   0,
				TotalIssues:     0,
				AssignedIssues:  0,
				CompletedIssues: 0,
				OverdueIssues:   0,
			}, nil
		},
	}
	service := stats.NewStatsService(store)
	_, err := service.GetStats(context.Background(), stats.GetStatsInput{
		WorkspaceID: workspaceID.String(),
		AssigneeID:  assigneeID.String(),
	})
	require.NoError(t, err)
}

func TestGetWorkspaceStats_ValidationErrors(t *testing.T) {
	validWorkspaceID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	validAssigneeID := mustUUID(t, "22222222-2222-2222-2222-222222222222")

	tests := []struct {
		name    string
		input   stats.GetStatsInput
		wantErr error
	}{
		{
			name: "invalid workspace id",
			input: stats.GetStatsInput{
				WorkspaceID: "invalid UUID",
				AssigneeID:  validAssigneeID.String(),
			},
			wantErr: common.ErrInvalidWorkspaceID,
		},
		{
			name: "imvalid assignee id",
			input: stats.GetStatsInput{
				WorkspaceID: validWorkspaceID.String(),
				AssigneeID:  "invalid UUID",
			},
			wantErr: common.ErrInvalidAssigneeID,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockStatsStore{
				getWorkspaceStatsFn: func(ctx context.Context, arg db.GetWorkspaceStatsParams) (db.GetWorkspaceStatsRow, error) {
					t.Fatal("GetWorkspaceStats should not be called with invalid input")
					return db.GetWorkspaceStatsRow{}, nil
				},
			}
			service := stats.NewStatsService(store)
			got, err := service.GetStats(context.Background(), tt.input)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, stats.Stats{}, got)
		})
	}
}

func TestGetWorkspaceStats_StoreError(t *testing.T) {
	workspaceID := mustUUID(t, "11111111-1111-1111-1111-111111111111")
	assigneeID := mustUUID(t, "22222222-2222-2222-2222-222222222222")
	expectedErr := errors.New("generic error")

	store := &mockStatsStore{
		getWorkspaceStatsFn: func(ctx context.Context, arg db.GetWorkspaceStatsParams) (db.GetWorkspaceStatsRow, error) {
			return db.GetWorkspaceStatsRow{}, expectedErr
		},
	}
	service := stats.NewStatsService(store)
	_, err := service.GetStats(context.Background(), stats.GetStatsInput{
		WorkspaceID: workspaceID.String(),
		AssigneeID:  assigneeID.String(),
	})
	require.ErrorIs(t, err, expectedErr)
}
