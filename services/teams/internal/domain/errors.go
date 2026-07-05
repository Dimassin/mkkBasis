package domain

import "errors"

var (
	ErrUserNotFound            = errors.New("user not found")
	ErrTeamNotFound            = errors.New("team not found")
	ErrForbidden               = errors.New("forbidden")
	ErrAlreadyMember           = errors.New("user is already a team member")
	ErrCircuitOpen             = errors.New("circuit breaker open")
	ErrEmailServiceUnavailable = errors.New("email service unavailable")
	ErrInternalServer          = errors.New("internal server error")
)
