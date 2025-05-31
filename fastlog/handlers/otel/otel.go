package otel

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"github.com/agoda-com/opentelemetry-go/otelslog"
	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs"
	sdk "github.com/agoda-com/opentelemetry-logs-go/sdk/logs"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/InsideGallery/core/fastlog/handlers"
	"github.com/InsideGallery/core/server/instance"
)

const OutKind = "otel"

func init() {
	handlers.RegisterHandler(OutKind,
		Handler(context.Background()),
	)
}

var (
	loggerProvider *LoggerProvider
	options        = &otelslog.HandlerOptions{}
	mu             sync.Mutex
)

func Default(ctx context.Context) *LoggerProvider {
	mu.Lock()
	defer mu.Unlock()

	if loggerProvider == nil {
		loggerProvider, options = NewProvider(ctx)
	}

	return loggerProvider
}

func Handler(ctx context.Context) slog.Handler {
	mu.Lock()
	defer mu.Unlock()

	if loggerProvider == nil {
		loggerProvider, options = NewProvider(ctx)
	}

	return otelslog.NewOtelHandler(loggerProvider, options)
}

type LoggerProvider struct {
	ctx context.Context
	*sdk.LoggerProvider
	TracerProvider *sdktrace.TracerProvider
}

func NewProvider(ctx context.Context) (*LoggerProvider, *otelslog.HandlerOptions) {
	cfg, err := GetConfigFromEnv()
	if err != nil {
		slog.Default().Error("Error get otel logger provider config", "err", err)
		return nil, nil
	}

	logExporter, err := otlplogs.NewExporter(ctx)
	if err != nil {
		slog.Default().Error("Error get otel logger provider exporter", "err", err)
		return nil, nil
	}

	exp, err := otlptracegrpc.New(ctx)
	if err != nil {
		slog.Default().Error("Error get otlptracegrpc logger", "err", err)
		return nil, nil
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(newResource(cfg)),
	)

	otel.SetTracerProvider(tracerProvider)

	return &LoggerProvider{
		ctx: ctx,
		LoggerProvider: sdk.NewLoggerProvider(
			sdk.WithBatcher(logExporter),
			sdk.WithResource(newResource(cfg)),
		),
		TracerProvider: tracerProvider,
	}, cfg.GetOptions()
}

func (l *LoggerProvider) Shutdown() {
	if l.ctx != nil {
		if l.LoggerProvider != nil {
			err := l.LoggerProvider.Shutdown(l.ctx)
			if err != nil {
				slog.Default().Error("Error call shutdown of logger provider", "err", err)
			}
		}

		if l.TracerProvider != nil {
			if err := l.TracerProvider.Shutdown(l.ctx); err != nil {
				slog.Default().Error("Error call shutdown of tracer provider", "err", err)
			}
		}
	}
}

func (l *LoggerProvider) Tracer(
	ctx context.Context, name, spanName string, kind trace.SpanKind,
) (context.Context, trace.Span) {
	tracer := otel.Tracer(name)
	return tracer.Start(ctx, spanName, trace.WithSpanKind(kind))
}

func (l *LoggerProvider) TracerEnd(span trace.Span) {
	span.End(trace.WithStackTrace(true))
}

func (l *LoggerProvider) TracerWrapper(
	ctx context.Context,
	name, spanName string,
	kind trace.SpanKind,
	fn func(ctx context.Context, span trace.Span),
) {
	ctx, span := l.Tracer(ctx, name, spanName, kind)
	defer l.TracerEnd(span)

	if fn != nil {
		fn(ctx, span)
	}
}

// configure common attributes for all logs
func newResource(cfg *Config) *resource.Resource {
	hostName, _ := os.Hostname()

	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(cfg.ServiceName),
		semconv.ServiceVersion(cfg.ServiceVersion),
		semconv.HostName(hostName),
		semconv.ServiceNamespace(cfg.Namespace),
		semconv.ServiceInstanceID(instance.GetInstanceID()),
	)
}
