package email

import (
	"context"
	"errors"
	"time"

	"teams/internal/adapter/circuitbreaker"
	"teams/internal/ports"
)

type CircuitBreakerService struct {
	breaker *circuitbreaker.Breaker
	inner   ports.EmailService
}

func NewCircuitBreakerService(inner ports.EmailService) *CircuitBreakerService {
	return &CircuitBreakerService{
		breaker: circuitbreaker.New(3, 30*time.Second),
		inner:   inner,
	}
}

func (s *CircuitBreakerService) SendTeamInvite(ctx context.Context, email string, teamID, inviterID int) error {
	err := s.breaker.Call(func() error {
		return s.inner.SendTeamInvite(ctx, email, teamID, inviterID)
	})
	if errors.Is(err, circuitbreaker.ErrOpen) {
		return circuitbreaker.ErrOpen
	}
	return err
}
