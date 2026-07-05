package ports

import (
	"auth/internal/domain"
	"context"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error

	FindByEmail(ctx context.Context, email string) (*domain.User, error)

	FindByID(ctx context.Context, id string) (*domain.User, error)
}
