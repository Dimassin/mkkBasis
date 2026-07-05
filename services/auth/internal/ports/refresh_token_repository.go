package ports

import (
	"context"
	"time"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, userID, token string, expiresAt time.Time) error
	FindByToken(ctx context.Context, token string) (userID string, err error)
	Revoke(ctx context.Context, token string) error
	RevokeAllForUser(ctx context.Context, userID string) error
}
