package ports

import (
	"context"
	"tasks/internal/domain"
)

type ReportRepository interface {
	GetTeamStats(ctx context.Context) ([]*domain.TeamStats, error)
	GetTopCreatorsByTeam(ctx context.Context) ([]*domain.TopCreator, error)
	GetInvalidAssigneeTasks(ctx context.Context) ([]*domain.InvalidAssigneeTask, error)
}

type ReportUsecase interface {
	GetTeamStats(ctx context.Context) ([]*domain.TeamStats, error)
	GetTopCreatorsByTeam(ctx context.Context) ([]*domain.TopCreator, error)
	GetInvalidAssigneeTasks(ctx context.Context) ([]*domain.InvalidAssigneeTask, error)
}
