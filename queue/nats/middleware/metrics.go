package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/InsideGallery/core/metrics"
	"github.com/InsideGallery/core/queue/nats/subscriber"
)

const (
	Count         = "nats.count"
	ContentLength = "nats.content_length"
	Latency       = "nats.duration"
)

type Metrics struct {
	client *metrics.Client
}

// CreateMeasuresWithClient creates NATS metrics middleware state from an explicit client.
func CreateMeasuresWithClient(client *metrics.Client) *Metrics {
	return &Metrics{client: client}
}

// CreateMeasures creates NATS metrics middleware state from the package-level metrics default.
func CreateMeasures() *Metrics {
	return CreateMeasuresWithClient(metrics.Default()) //nolint:staticcheck // legacy wrapper reads compatibility default
}

// SubMetrics implement Middleware interface
type SubMetrics struct {
	*Metrics
}

func NewMetrics(m *Metrics) *SubMetrics {
	return &SubMetrics{Metrics: m}
}

//nolint:staticcheck // legacy middleware keeps NATS handler shim
func (t *SubMetrics) Call(next subscriber.MsgHandler) subscriber.MsgHandler {
	return func(ctx context.Context, msg *nats.Msg) error {
		var err error

		defer func(start time.Time) {
			if ctx.Err() != nil {
				err = ctx.Err()
			}

			tags := []string{
				"error:" + boolTag(err != nil),
				"subject:" + msg.Subject,
			}

			recordCount(t.client, Count, 1, tags)
			recordCount(t.client, ContentLength, int64(len(msg.Data)), tags)
			recordDistribution(t.client, Latency, float64(time.Since(start).Milliseconds()), tags)
		}(time.Now())

		return next(ctx, msg)
	}
}

func boolTag(value bool) string {
	if value {
		return "true"
	}

	return "false"
}

func recordCount(client *metrics.Client, name string, value int64, tags []string) {
	if err := client.Count(name, value, tags); err != nil {
		slog.Default().Warn("record nats metric failed", "metric", name, "err", err)
	}
}

func recordDistribution(client *metrics.Client, name string, value float64, tags []string) {
	if err := client.Distribution(name, value, tags); err != nil {
		slog.Default().Warn("record nats metric failed", "metric", name, "err", err)
	}
}
