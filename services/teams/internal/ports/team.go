package ports

import (
	"context"
	"teams/internal/domain"
)

type TeamRepository interface {
	Create(ctx context.Context, team *domain.Team) error
	AddMember(ctx context.Context, teamID, userID int, role string) error
	GetUserTeams(ctx context.Context, userID int) ([]*domain.Team, error)
	TeamExists(ctx context.Context, teamID int) (bool, error)
	GetMemberRole(ctx context.Context, teamID, userID int) (string, error)
}

type TeamUsecase interface {
	CreateTeam(ctx context.Context, userID int, req *domain.CreateTeamRequest) (*domain.TeamResponse, error)
	GetUserTeams(ctx context.Context, userID int) ([]*domain.TeamResponse, error)
	InviteMember(ctx context.Context, inviterID, teamID int, req *domain.InviteMemberRequest) (*domain.InviteMemberResponse, error)
}
