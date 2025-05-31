package ticker

import (
	"context"
	"time"

	"go.uber.org/atomic"
)

// ExecuteWithDelay execute method with delay
type ExecuteWithDelay struct {
	close  chan struct{}
	active *atomic.Bool
}

// NewExecuteWithDelay return new execute in delay
func NewExecuteWithDelay() *ExecuteWithDelay {
	return &ExecuteWithDelay{
		close:  make(chan struct{}),
		active: atomic.NewBool(false),
	}
}

// IsActive return true if active
func (e *ExecuteWithDelay) IsActive() bool {
	return e.active.Load()
}

// Start start timer
func (e *ExecuteWithDelay) Start(ctx context.Context, h func(ctx context.Context), d time.Duration) {
	if e.IsActive() {
		return
	}

	e.active.Store(true)

	go func(ctx context.Context, e *ExecuteWithDelay, h func(ctx context.Context), d time.Duration) {
		tickTimer := time.NewTimer(d)

		for {
			select {
			case <-e.close:
				e.active.Store(false)

				return
			case <-tickTimer.C:
				h(ctx)
				e.active.Store(false)

				return
			}
		}
	}(ctx, e, h, d)
}

// Stop stop executing
func (e *ExecuteWithDelay) Stop() bool {
	if !e.IsActive() {
		return false
	}

	e.close <- struct{}{}

	return true
}
