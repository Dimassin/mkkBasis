package jwtadapter

import (
	"auth/internal/domain"
	"auth/internal/ports"
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	secretKey      string
	accessTokenTTL time.Duration
}

func (m *JWTManager) GenerateAccessToken(ctx context.Context, user *domain.User) (string, error) {
	claims := &domain.UserClaims{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secretKey))
}

func (m *JWTManager) ValidateToken(ctx context.Context, token string) (*domain.UserClaims, error) {
	claims := &domain.UserClaims{}

	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	return claims, nil
}

func NewJWTManager(secretKey string, accessTTL time.Duration) ports.TokenManager {
	return &JWTManager{
		secretKey:      secretKey,
		accessTokenTTL: accessTTL,
	}
}
