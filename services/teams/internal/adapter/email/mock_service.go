package email

import (
	"context"
	"errors"
	"os"
	"time"
)

var ErrUnavailable = errors.New("email service unavailable")

type MockService struct{}

func NewMockService() *MockService {
	return &MockService{}
}

func (s *MockService) SendTeamInvite(ctx context.Context, email string, teamID, inviterID int) error {
	if os.Getenv("MOCK_EMAIL_FAIL") == "true" {
		return ErrUnavailable
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(10 * time.Millisecond):
		return nil
	}
}
