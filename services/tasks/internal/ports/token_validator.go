package ports

import (
	"context"
	"tasks/internal/domain"
)

type TokenValidator interface {
	ValidateToken(ctx context.Context, token string) (*domain.UserClaims, error)
}
