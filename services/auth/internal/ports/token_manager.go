package ports

import (
	"auth/internal/domain"
	"context"
)

type TokenManager interface {
	GenerateAccessToken(ctx context.Context, user *domain.User) (string, error)
	ValidateToken(ctx context.Context, token string) (*domain.UserClaims, error)
}
