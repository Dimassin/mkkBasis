package ports

import (
	"context"
	"teams/internal/domain"
)

type TokenValidator interface {
	ValidateToken(ctx context.Context, token string) (*domain.UserClaims, error)
}
