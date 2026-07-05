package hasher

import (
	"auth/internal/ports"
	"context"

	"golang.org/x/crypto/bcrypt"
)

type BcryptHasher struct{}

func NewBcryptHasher() ports.PasswordHasher {
	return &BcryptHasher{}
}

func (h *BcryptHasher) Hash(ctx context.Context, password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (h *BcryptHasher) Compare(ctx context.Context, hashedPassword, plainPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}
