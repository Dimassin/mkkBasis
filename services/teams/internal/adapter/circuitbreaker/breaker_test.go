package circuitbreaker_test

import (
	"errors"
	"testing"
	"time"

	"teams/internal/adapter/circuitbreaker"
)

func TestBreakerOpensAfterFailures(t *testing.T) {
	breaker := circuitbreaker.New(2, time.Minute)
	errFail := errors.New("fail")

	_ = breaker.Call(func() error { return errFail })
	_ = breaker.Call(func() error { return errFail })

	err := breaker.Call(func() error { return nil })
	if !errors.Is(err, circuitbreaker.ErrOpen) {
		t.Fatalf("expected circuit open, got %v", err)
	}
}
