package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

var ErrOpen = errors.New("circuit breaker open")

type Breaker struct {
	maxFailures  int
	resetTimeout time.Duration

	mu       sync.Mutex
	failures int
	state    string
	openedAt time.Time
}

func New(maxFailures int, resetTimeout time.Duration) *Breaker {
	return &Breaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        "closed",
	}
}

func (b *Breaker) Call(fn func() error) error {
	if err := b.beforeCall(); err != nil {
		return err
	}

	err := fn()
	b.afterCall(err)
	return err
}

func (b *Breaker) beforeCall() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.state == "open" {
		if time.Since(b.openedAt) >= b.resetTimeout {
			b.state = "half-open"
			return nil
		}
		return ErrOpen
	}

	return nil
}

func (b *Breaker) afterCall(err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err != nil {
		b.failures++
		b.openedAt = time.Now()
		if b.failures >= b.maxFailures {
			b.state = "open"
		}
		return
	}

	b.failures = 0
	b.state = "closed"
}
