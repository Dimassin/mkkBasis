package domain

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrWeakPassword       = errors.New("password must be at least 6 characters")
	ErrInternalServer     = errors.New("internal server error")
)
