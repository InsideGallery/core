package otel

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/agoda-com/opentelemetry-go/otelslog"
	"github.com/agoda-com/opentelemetry-logs-go/exporters/otlp/otlplogs"
	sdklog "github.com/agoda-com/opentelemetry-logs-go/sdk/logs"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/InsideGallery/core/fastlog/handlers"
)

// OutKind is the registry key for the otel handler.
const OutKind = "otel"

func init() {
	handlers.RegisterHandlerFunc(OutKind, newHandler)
}

// LoggerProvider holds OTEL log and trace providers.
type LoggerProvider struct {
	logProvider   *sdklog.LoggerProvider
	traceProvider *sdktrace.TracerProvider
}

var (
	provider *LoggerProvider
	once     sync.Once
)

// Default returns the singleton LoggerProvider, initializing it on first call.
func Default(ctx context.Context) *LoggerProvider {
	once.Do(func() {
		p, err := NewProvider(ctx)
		if err != nil {
			slog.Error("Failed to init OTEL provider", "err", err)

			provider = &LoggerProvider{}

			return
		}

		provider = p
	})

	return provider
}

// NewProvider creates OTEL log and trace providers from env config.
func NewProvider(ctx context.Context) (*LoggerProvider, error) {
	cfg, err := getConfigFromEnv()
	if err != nil {
		return nil, err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("get hostname: %w", err)
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(cfg.ServiceName),
		semconv.ServiceVersion(cfg.ServiceVersion),
		semconv.ServiceNamespace(cfg.Namespace),
		semconv.ServiceInstanceID(hostname),
	)

	logExporter, err := otlplogs.NewExporter(ctx)
	if err != nil {
		return nil, fmt.Errorf("create otel log exporter: %w", err)
	}

	logProv := sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithBatcher(logExporter),
	)

	traceExporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("create otel trace exporter: %w", err)
	}

	traceProv := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(traceExporter),
	)

	otel.SetTracerProvider(traceProv)

	return &LoggerProvider{
		logProvider:   logProv,
		traceProvider: traceProv,
	}, nil
}

// Shutdown gracefully shuts down both providers.
func (p *LoggerProvider) Shutdown() {
	ctx := context.Background()

	var errs []error

	if p.logProvider != nil {
		if err := p.logProvider.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if p.traceProvider != nil {
		if err := p.traceProvider.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if err := errors.Join(errs...); err != nil {
		slog.Error("OTEL shutdown errors", "err", err)
	}
}

// Tracer starts a span from the global OpenTelemetry tracer provider.
func (p *LoggerProvider) Tracer(
	ctx context.Context,
	name string,
	spanName string,
	kind trace.SpanKind,
) (context.Context, trace.Span) {
	tracer := otel.Tracer(name)

	return tracer.Start(ctx, spanName, trace.WithSpanKind(kind))
}

// TracerEnd ends a span with stack trace recording enabled.
func (p *LoggerProvider) TracerEnd(span trace.Span) {
	span.End(trace.WithStackTrace(true))
}

// TracerWrapper runs fn inside an OpenTelemetry span.
func (p *LoggerProvider) TracerWrapper(
	ctx context.Context,
	name string,
	spanName string,
	kind trace.SpanKind,
	fn func(ctx context.Context, span trace.Span),
) {
	ctx, span := p.Tracer(ctx, name, spanName, kind)
	defer p.TracerEnd(span)

	if fn != nil {
		fn(ctx, span)
	}
}

func newHandler() (slog.Handler, error) {
	cfg, err := getConfigFromEnv()
	if err != nil {
		return nil, err
	}

	provider := Default(context.Background())
	if provider == nil || provider.logProvider == nil {
		return nil, errors.New("otel provider is not initialized")
	}

	return otelslog.NewOtelHandler(provider.logProvider, &otelslog.HandlerOptions{
		Level: cfg.Level,
	}), nil
}
