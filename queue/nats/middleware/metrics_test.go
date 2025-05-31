package middleware

import (
	"context"
	"log/slog"
	"testing"

	_ "github.com/InsideGallery/core/fastlog/handlers/otel"
	_ "github.com/InsideGallery/core/fastlog/handlers/stderr"

	"github.com/nats-io/nats.go"

	"github.com/InsideGallery/core/fastlog"
	"github.com/InsideGallery/core/fastlog/handlers/otel"
	"github.com/InsideGallery/core/fastlog/metrics"
	"github.com/InsideGallery/core/testutils"
)

func TestMetrics(t *testing.T) {
	ctx := context.Background()
	fastlog.SetupDefaultLog()
	m, err := metrics.Default(ctx)
	testutils.Equal(t, err, nil)
	defer m.Shutdown()
	defer otel.Default(ctx).Shutdown()

	mm := NewMetrics(CreateMeasures())
	err = mm.Call(func(ctx context.Context, _ *nats.Msg) error {
		slog.Default().ErrorContext(ctx, "Log message of metrics collect")
		return nil
	})(ctx, &nats.Msg{
		Subject: "test-metric",
	})

	testutils.Equal(t, err, nil)
}
