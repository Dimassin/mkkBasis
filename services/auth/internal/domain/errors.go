package domain

import "errors"

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrWeakPassword      = errors.New("password must be at least 6 characters")
	ErrInternalServer    = errors.New("internal server error")
)
