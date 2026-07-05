package usecase

import (
	"auth/internal/domain"
	"auth/internal/ports"
	"context"
	"strconv"
)

type TeamUsecase struct {
	teamRepo ports.TeamRepository
	userRepo ports.UserRepository
}

func NewTeamUsecase(teamRepo ports.TeamRepository, userRepo ports.UserRepository) *TeamUsecase {
	return &TeamUsecase{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

func (uc *TeamUsecase) CreateTeam(ctx context.Context, userID int, req *domain.CreateTeamRequest) (*domain.TeamResponse, error) {
	// Проверяем, существует ли пользователь
	_, err := uc.userRepo.FindByID(ctx, strconv.Itoa(userID))
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	// Создаем команду
	team := &domain.Team{
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   userID,
	}

	if err := uc.teamRepo.Create(ctx, team); err != nil {
		return nil, err
	}

	// Добавляем создателя как owner
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
