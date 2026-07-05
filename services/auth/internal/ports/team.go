package ports

import (
	"auth/internal/domain"
	"context"
)

type TeamRepository interface {
	Create(ctx context.Context, team *domain.Team) error
	AddMember(ctx context.Context, teamID, userID int, role string) error
}

type TeamUsecase interface {
	CreateTeam(ctx context.Context, userID int, req *domain.CreateTeamRequest) (*domain.TeamResponse, error)
}
