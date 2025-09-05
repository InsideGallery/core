package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/InsideGallery/core/fastlog/metrics"
	"github.com/InsideGallery/core/queue/nats/subscriber"
)

const (
	Count         = "nats.count"
	ContentLength = "nats.content_length"
	Latency       = "nats.duration"

	Subject = attribute.Key("subject")
	IsError = attribute.Key("error")
)

type Metrics struct {
	counters       map[string]metric.Int64Counter
	valueRecorders map[string]metric.Float64Histogram
}

func CreateMeasures() *Metrics {
	cfg, err := metrics.GetConfigFromEnv()
	if err != nil {
		slog.Default().Error("Error getting config of metrics", "err", err)
	}

	m := otel.GetMeterProvider().Meter(cfg.Namespace)

	counters := make(map[string]metric.Int64Counter)
	valueRecorders := make(map[string]metric.Float64Histogram)

	counter, err := m.Int64Counter(Count)
	if err != nil {
		slog.Default().Error("Error getting int64 counter", "err", err)
	}

	requestBytesCounter, err := m.Int64Counter(ContentLength)
	if err != nil {
		slog.Default().Error("Error getting int64 histogram", "err", err)
	}

	serverLatencyMeasure, err := m.Float64Histogram(Latency)
	if err != nil {
		slog.Default().Error("Error getting float64 histogram", "err", err)
	}

	counters[Count] = counter
	counters[ContentLength] = requestBytesCounter
	valueRecorders[Latency] = serverLatencyMeasure

	return &Metrics{
		counters:       counters,
		valueRecorders: valueRecorders,
	}
}

// SubMetrics implement Middleware interface
type SubMetrics struct {
	*Metrics
}

func NewMetrics(m *Metrics) *SubMetrics {
	return &SubMetrics{Metrics: m}
}

func (t *SubMetrics) Call(next subscriber.MsgHandler) subscriber.MsgHandler {
	return func(ctx context.Context, msg *nats.Msg) error {
		var err error

		defer func(start time.Time) {
			if ctx.Err() != nil {
				err = ctx.Err()
			}

			attr := []attribute.KeyValue{
				IsError.Bool(err != nil),
				Subject.String(msg.Subject),
			}

			t.counters[Count].Add(ctx, 1, metric.WithAttributes(attr...))
			t.counters[ContentLength].Add(ctx, int64(len(msg.Data)), metric.WithAttributes(attr...))
			t.valueRecorders[Latency].Record(ctx, float64(time.Since(start).Milliseconds()), metric.WithAttributes(attr...))
		}(time.Now())

		return next(ctx, msg)
	}
}
