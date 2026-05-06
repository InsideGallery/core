package middlewares

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
)

var (
	dur   time.Duration
	count int
	mu    sync.RWMutex
)

// Timing calculate time of request
func Timing(next fiber.Handler) fiber.Handler {
	return func(c fiber.Ctx) error {
		st := time.Now()

		err := next(c)
		if err != nil {
			return err
		}

		go func() {
			mu.Lock()
			defer mu.Unlock()

			dur += time.Since(st)
			count++
		}()

		return nil
	}
}

// TimingStats returns current average response duration and resets counters.
func TimingStats() (time.Duration, int) {
	mu.Lock()
	defer mu.Unlock()

	d := dur
	c := count

	dur = 0
	count = 0

	if c == 0 {
		return 0, 0
	}

	return d / time.Duration(c), c
}

// StartTimingReporter starts a goroutine that logs average response time every interval.
// Cancel the context to stop it.
func StartTimingReporter(ctx context.Context) {
	t := time.NewTicker(time.Minute)

	go func() {
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				avg, c := TimingStats()
				if c > 0 {
					slog.Default().Info("Current average response", "duration", avg.String(), "count", c)
				}
			}
		}
	}()
}
