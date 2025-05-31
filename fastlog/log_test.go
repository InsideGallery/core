//go:build local_test
// +build local_test

package fastlog

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"go.opentelemetry.io/otel/trace"

	_ "github.com/InsideGallery/core/fastlog/handlers/logfile"
	_ "github.com/InsideGallery/core/fastlog/handlers/logstash"
	_ "github.com/InsideGallery/core/fastlog/handlers/nop"
	"github.com/InsideGallery/core/fastlog/handlers/otel"
	_ "github.com/InsideGallery/core/fastlog/handlers/otel"
	_ "github.com/InsideGallery/core/fastlog/handlers/stderr"
	_ "github.com/InsideGallery/core/fastlog/handlers/stdout"
	"github.com/InsideGallery/core/fastlog/middlewares"
)

func TestLog(t *testing.T) {
	SetupDefaultLog(middlewares.NewGDPRMiddleware())
	defer otel.Default(context.Background()).Shutdown()

	slog.Default().
		With(
			slog.Group("user",
				slog.String("id", "user-123"),
				slog.String("email", "user-123"),
				slog.Time("created_at", time.Now()),
			),
		).
		With("environment", "dev").
		With("password", "maxim").
		Error("A message",
			slog.String("foo", "bar"),
			slog.Any("error", fmt.Errorf("an error")))
}

func TestTrace(t *testing.T) {
	ctx := context.Background()
	SetupDefaultLog(middlewares.NewGDPRMiddleware())
	defer otel.Default(ctx).Shutdown()

	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{0x03},
		SpanID:  trace.SpanID{0x03},
	})
	ctx = trace.ContextWithRemoteSpanContext(ctx, sc)

	ctx, span := otel.Default(ctx).Tracer(ctx, "default", "span-context", trace.SpanKindClient)
	slog.Default().
		ErrorContext(ctx, "New log message",
			slog.String("foo", "bar"),
			slog.Any("error", fmt.Errorf("an error")))
	slog.Default().
		ErrorContext(ctx, "New log message2",
			slog.String("foo", "bar"),
			slog.Any("error", fmt.Errorf("an error")))
	otel.Default(ctx).TracerEnd(span)

	ctx, span = otel.Default(ctx).Tracer(ctx, "default", "span-context2", trace.SpanKindClient)
	slog.Default().
		ErrorContext(ctx, "New log message3",
			slog.String("foo", "bar"),
			slog.Any("error", fmt.Errorf("an error")))
	otel.Default(ctx).TracerEnd(span)
}
