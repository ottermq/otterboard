package stats

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ottermq/otterboard/src/backend/internal/common"
	"github.com/ottermq/otterboard/src/backend/internal/db"
)

type StatsStore interface {
	GetWorkspaceStats(ctx context.Context, arg db.GetWorkspaceStatsParams) (db.GetWorkspaceStatsRow, error)
}

type StatsService struct {
	store StatsStore
}

type Stats struct {
	TotalProjects   int32
	TotalIssues     int32
	AssignedIssues  int32
	CompletedIssues int32
	OverdueIssues   int32
}

type GetStatsInput struct {
	WorkspaceID string
	AssigneeID  string
}

func NewStatsService(store StatsStore) *StatsService {
	return &StatsService{
		store: store,
	}
}

func (s *StatsService) GetStats(ctx context.Context, input GetStatsInput) (Stats, error) {
	var workspaceUUID pgtype.UUID
	if err := workspaceUUID.Scan(input.WorkspaceID); err != nil {
		return Stats{}, common.ErrInvalidWorkspaceID
	}
	var assigneeUUID pgtype.UUID
	if err := assigneeUUID.Scan(input.AssigneeID); err != nil {
		return Stats{}, common.ErrInvalidAssigneeID
	}

	statsRow, err := s.store.GetWorkspaceStats(ctx, db.GetWorkspaceStatsParams{
		WorkspaceID: workspaceUUID,
		AssigneeID:  assigneeUUID,
	})
	if err != nil {
		return Stats{}, err
	}
	return mapToStatsDomain(statsRow), nil
}

func mapToStatsDomain(statsRow db.GetWorkspaceStatsRow) Stats {
	return Stats{
		TotalProjects:   statsRow.TotalProjects,
		TotalIssues:     statsRow.TotalIssues,
		AssignedIssues:  statsRow.AssignedIssues,
		CompletedIssues: statsRow.CompletedIssues,
		OverdueIssues:   statsRow.OverdueIssues,
	}
}
