package throughput

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/InsideGallery/core/server/instance"
)

const (
	Tier0 = iota
	Tier1
	Tier2
	Tier3

	Tier0RPS uint64 = 10
	Tier1RPS uint64 = 50
	Tier2RPS uint64 = 250
	Tier3RPS uint64 = 1250

	Tier0RPM uint64 = 1000000
	Tier1RPM uint64 = 5000000
	Tier2RPM uint64 = 25000000
	Tier3RPM uint64 = 125000000
)

var tiers = map[int]uint64{
	Tier0: Tier0RPS,
	Tier1: Tier1RPS,
	Tier2: Tier2RPS,
	Tier3: Tier3RPS,
}

var tiersM = map[int]uint64{
	Tier0: Tier0RPM,
	Tier1: Tier1RPM,
	Tier2: Tier2RPM,
	Tier3: Tier3RPM,
}

func GetRPS(tier int) uint64 {
	return tiers[tier]
}

func GetRPM(tier int) uint64 {
	return tiersM[tier]
}

type Throughput struct {
	ctx     context.Context
	storage Storage
}

func New(ctx context.Context, storage Storage) *Throughput {
	return &Throughput{
		ctx:     ctx,
		storage: storage,
	}
}

func (t *Throughput) Validate(name string) bool {
	crd := t.storage.RPS(name)
	if crd >= GetRPS(t.storage.Tier(name)) {
		return false
	}

	crm := t.storage.RPM(name)
	if crm >= GetRPM(t.storage.Tier(name)) {
		return false
	}

	t.storage.Incr(name)

	return true
}

func (t *Throughput) Loop() {
	tk := time.NewTicker(time.Second)

	for {
		select {
		case <-t.ctx.Done():
			return
		case <-tk.C:
			t.storage.Reset()
		}
	}
}

func (t *Throughput) Middleware(parameter string) func(next fiber.Handler) fiber.Handler {
	return func(next fiber.Handler) fiber.Handler {
		return func(c *fiber.Ctx) error {
			st := time.Now()
			defer func() {
				latency := time.Since(st)
				if latency > time.Second {
					slog.Warn("Latency is higher of second",
						"latency", latency.String(),
						"siid", instance.GetShortInstanceID(),
					)
				}
			}()

			val := c.Locals(parameter).(string)
			if !t.Validate(val) {
				slog.Warn("Too many requests",
					"tier", t.storage.Tier(val),
					"max rps", t.storage.RPS(val),
					"max rpm", t.storage.RPM(val),
					"siid", instance.GetShortInstanceID(),
				)

				c.Status(http.StatusTooManyRequests)

				_, err := c.Write([]byte{})

				return err
			}

			return next(c)
		}
	}
}
