package domain

import "github.com/golang-jwt/jwt/v5"

type UserClaims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}
