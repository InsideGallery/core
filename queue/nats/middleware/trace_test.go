package middleware

import (
	"context"
	"log/slog"
	"testing"

	_ "github.com/InsideGallery/core/fastlog/handlers/stderr"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/trace"

	"github.com/InsideGallery/core/fastlog"
	"github.com/InsideGallery/core/fastlog/handlers/otel"
	"github.com/InsideGallery/core/fastlog/middlewares"
	"github.com/InsideGallery/core/queue/nats/natsprop"
	"github.com/InsideGallery/core/testutils"
)

func TestTrace(t *testing.T) {
	ctx := context.Background()
	fastlog.SetupDefaultLog(middlewares.NewGDPRMiddleware())
	defer otel.Default(ctx).Shutdown()

	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{0x03},
		SpanID:  trace.SpanID{0x03},
	})
	ctx = trace.ContextWithRemoteSpanContext(ctx, sc)

	tr := NewTracer()
	err := tr.Call(func(ctx context.Context, msg *nats.Msg) error {
		t.Helper()

		slog.Default().ErrorContext(ctx, "Log message with external trace id")
		spanContext := natsprop.Extract(ctx, msg)
		testutils.Equal(t, spanContext.TraceID(), trace.TraceID{0x03})

		return nil
	})(ctx, &nats.Msg{})
	testutils.Equal(t, err, nil)
}
