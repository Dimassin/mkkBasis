package usecase

import (
	"auth/internal/domain"
	"auth/internal/ports"
	"context"
)

type AuthUsecase struct {
	userRepo ports.UserRepository
	tokenMgr ports.TokenManager
	hasher   ports.PasswordHasher
}

func NewAuthUsecase(
	userRepo ports.UserRepository,
	tokenMgr ports.TokenManager,
	hasher ports.PasswordHasher,
) *AuthUsecase {
	return &AuthUsecase{
		userRepo: userRepo,
		tokenMgr: tokenMgr,
		hasher:   hasher,
	}
}

func (uc *AuthUsecase) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.AuthResponse, error) {
	existing, _ := uc.userRepo.FindByEmail(ctx, req.Email)
	if existing != nil {
		return nil, domain.ErrUserAlreadyExists
	}
	if len(req.Password) < 6 {
		return nil, domain.ErrWeakPassword
	}

	hashedPassword, err := uc.hasher.Hash(ctx, req.Password)
	if err != nil {
		return nil, domain.ErrInternalServer
	}

	user := &domain.User{
		Email:    req.Email,
		Password: hashedPassword,
		Username: req.Username,
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	accessToken, err := uc.tokenMgr.GenerateAccessToken(ctx, user)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		AccessToken: accessToken,
		UserID:      user.ID,
		Email:       user.Email,
		Username:    user.Username,
	}, nil
}

func (uc *AuthUsecase) Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error) {
	user, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}
	if user == nil {
		return nil, domain.ErrInvalidCredentials
	}

	if err := uc.hasher.Compare(ctx, user.Password, req.Password); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	accessToken, err := uc.tokenMgr.GenerateAccessToken(ctx, user)
	if err != nil {
		return nil, domain.ErrInternalServer
	}

	return &domain.AuthResponse{
		AccessToken: accessToken,
		UserID:      user.ID,
		Email:       user.Email,
		Username:    user.Username,
	}, nil
}
