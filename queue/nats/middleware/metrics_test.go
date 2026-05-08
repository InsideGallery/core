package middleware

import (
	"context"
	"log/slog"
	"testing"

	"github.com/nats-io/nats.go"

	"github.com/InsideGallery/core/fastlog"
	"github.com/InsideGallery/core/metrics"
	"github.com/InsideGallery/core/testutils"
)

func TestMetrics(t *testing.T) {
	ctx := context.Background()

	handle, err := fastlog.SetupDefaultLogger(&fastlog.Config{
		Outputs: []string{"stderr:json"},
		Level:   slog.LevelInfo,
	})
	if err != nil {
		t.Fatalf("setup default logger: %v", err)
	}
	defer handle.Close()

	defaultHandle := metrics.InstallDefault(nil)
	defer defaultHandle.Close()

	mm := NewMetrics(CreateMeasures())
	err = mm.Call(func(ctx context.Context, _ *nats.Msg) error {
		slog.Default().ErrorContext(ctx, "Log message of metrics collect")
		return nil
	})(ctx, &nats.Msg{
		Subject: "test-metric",
	})

	testutils.Equal(t, err, nil)
}
