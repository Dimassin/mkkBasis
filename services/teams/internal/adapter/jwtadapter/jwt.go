package jwtadapter

import (
	"context"
	"teams/internal/domain"
	"teams/internal/ports"

	"github.com/golang-jwt/jwt/v5"
)

type JWTValidator struct {
	secretKey string
}

func (m *JWTValidator) ValidateToken(ctx context.Context, token string) (*domain.UserClaims, error) {
	claims := &domain.UserClaims{}

	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.secretKey), nil
	})
	if err != nil {
		return nil, err
	}

	return claims, nil
}

func NewJWTValidator(secretKey string) ports.TokenValidator {
	return &JWTValidator{secretKey: secretKey}
}
