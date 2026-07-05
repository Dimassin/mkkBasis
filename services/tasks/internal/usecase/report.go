package usecase

import (
	"context"
	"tasks/internal/domain"
	"tasks/internal/ports"
)

type ReportUsecase struct {
	reportRepo ports.ReportRepository
}

func NewReportUsecase(reportRepo ports.ReportRepository) *ReportUsecase {
	return &ReportUsecase{reportRepo: reportRepo}
}

func (uc *ReportUsecase) GetTeamStats(ctx context.Context) ([]*domain.TeamStats, error) {
	return uc.reportRepo.GetTeamStats(ctx)
}

func (uc *ReportUsecase) GetTopCreatorsByTeam(ctx context.Context) ([]*domain.TopCreator, error) {
	return uc.reportRepo.GetTopCreatorsByTeam(ctx)
}

func (uc *ReportUsecase) GetInvalidAssigneeTasks(ctx context.Context) ([]*domain.InvalidAssigneeTask, error) {
	return uc.reportRepo.GetInvalidAssigneeTasks(ctx)
}
