package ports

import "context"

type PasswordHasher interface {
	Hash(ctx context.Context, password string) (string, error)

	Compare(ctx context.Context, hashedPassword, plainPassword string) error
}
