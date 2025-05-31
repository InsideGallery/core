package middlewares

import (
	"log/slog"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

var (
	dur   time.Duration
	count int
	mu    sync.RWMutex
)

func init() {
	t := time.NewTicker(time.Minute)
	go func() {
		for range t.C {
			mu.RLock()
			if count == 0 {
				mu.RUnlock()
				continue
			}

			slog.Default().Info("Current average response", "duration", (dur / time.Duration(count)).String())

			mu.RUnlock()
		}
	}()
}

// Timing calculate time of request
func Timing(next fiber.Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
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
