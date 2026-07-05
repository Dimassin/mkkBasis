package usecase

import (
	"context"
	"errors"
	"strconv"
	"teams/internal/adapter/circuitbreaker"
	"teams/internal/adapter/email"
	"teams/internal/domain"
	"teams/internal/ports"
)

type TeamUsecase struct {
	teamRepo     ports.TeamRepository
	userRepo     ports.UserRepository
	emailService ports.EmailService
}

func NewTeamUsecase(teamRepo ports.TeamRepository, userRepo ports.UserRepository, emailService ports.EmailService) *TeamUsecase {
	return &TeamUsecase{
		teamRepo:     teamRepo,
		userRepo:     userRepo,
		emailService: emailService,
	}
}

func (uc *TeamUsecase) CreateTeam(ctx context.Context, userID int, req *domain.CreateTeamRequest) (*domain.TeamResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, strconv.Itoa(userID))
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	team := &domain.Team{
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   userID,
	}

	if err := uc.teamRepo.Create(ctx, team); err != nil {
		return nil, err
	}

	if err := uc.teamRepo.AddMember(ctx, team.ID, userID, "owner"); err != nil {
		return nil, err
	}

	return &domain.TeamResponse{
		ID:          team.ID,
		Name:        team.Name,
		Description: team.Description,
		CreatedBy:   team.CreatedBy,
		CreatedAt:   team.CreatedAt,
	}, nil
}

func (uc *TeamUsecase) GetUserTeams(ctx context.Context, userID int) ([]*domain.TeamResponse, error) {
	teams, err := uc.teamRepo.GetUserTeams(ctx, userID)
	if err != nil {
		return nil, err
	}

	response := make([]*domain.TeamResponse, 0, len(teams))
	for _, team := range teams {
		response = append(response, &domain.TeamResponse{
			ID:          team.ID,
			Name:        team.Name,
			Description: team.Description,
			CreatedBy:   team.CreatedBy,
			CreatedAt:   team.CreatedAt,
		})
	}

	return response, nil
}

func (uc *TeamUsecase) InviteMember(ctx context.Context, inviterID, teamID int, req *domain.InviteMemberRequest) (*domain.InviteMemberResponse, error) {
	exists, err := uc.teamRepo.TeamExists(ctx, teamID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domain.ErrTeamNotFound
	}

	inviterRole, err := uc.teamRepo.GetMemberRole(ctx, teamID, inviterID)
	if err != nil {
		return nil, err
	}
	if inviterRole != "owner" && inviterRole != "admin" {
		return nil, domain.ErrForbidden
	}

	user, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	inviteeID, err := strconv.Atoi(user.ID)
	if err != nil {
		return nil, domain.ErrInternalServer
	}

	memberRole, err := uc.teamRepo.GetMemberRole(ctx, teamID, inviteeID)
	if err != nil {
		return nil, err
	}
	if memberRole != "" {
		return nil, domain.ErrAlreadyMember
	}

	role := "member"
	if err := uc.emailService.SendTeamInvite(ctx, user.Email, teamID, inviterID); err != nil {
		if errors.Is(err, circuitbreaker.ErrOpen) {
			return nil, domain.ErrCircuitOpen
		}
		if errors.Is(err, email.ErrUnavailable) {
			return nil, domain.ErrEmailServiceUnavailable
		}
		return nil, err
	}

	if err := uc.teamRepo.AddMember(ctx, teamID, inviteeID, role); err != nil {
		return nil, err
	}

	return &domain.InviteMemberResponse{
		TeamID: teamID,
		UserID: inviteeID,
		Email:  user.Email,
		Role:   role,
	}, nil
}
