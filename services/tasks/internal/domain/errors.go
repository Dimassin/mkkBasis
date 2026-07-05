package domain

import "errors"

var (
	ErrTaskNotFound    = errors.New("task not found")
	ErrTeamNotFound    = errors.New("team not found")
	ErrForbidden       = errors.New("forbidden")
	ErrInvalidStatus   = errors.New("invalid status")
	ErrInvalidPriority = errors.New("invalid priority")
	ErrInternalServer  = errors.New("internal server error")
)
