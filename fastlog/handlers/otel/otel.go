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
	// Deprecated: call Setup with an explicit registry and handler factory.
	err := Setup(handlers.DefaultRegistry(), legacyHandlerFactory(context.Background()))
	if err != nil {
		slog.Default().Error("setup otel handler", "err", err)
	}
}

var (
	loggerProvider *LoggerProvider
	options        = &otelslog.HandlerOptions{}
	mu             sync.Mutex
)

// Setup registers the OpenTelemetry handler factory on an explicit registry.
func Setup(registry *handlers.Registry, factory handlers.HandlerFactory) error {
	if registry == nil {
		return errors.New("handler registry is nil")
	}

	if factory == nil {
		return errors.New("handler factory is nil")
	}

	registry.RegisterHandlerFactory(OutKind, factory)

	return nil
}

// NewHandlerFactory creates a handler factory from explicit OpenTelemetry dependencies.
func NewHandlerFactory(provider *LoggerProvider, opts *otelslog.HandlerOptions) handlers.HandlerFactory {
	return func() slog.Handler {
		return NewHandler(provider, opts)
	}
}

func legacyHandlerFactory(ctx context.Context) handlers.HandlerFactory {
	return func() slog.Handler {
		return Handler(ctx)
	}
}

// Default returns the package-level OpenTelemetry logger provider.
//
// Deprecated: use NewProviderFromConfig and pass the provider explicitly.
func Default(ctx context.Context) *LoggerProvider {
	mu.Lock()
	defer mu.Unlock()

	if loggerProvider == nil {
		loggerProvider, options = NewProvider(ctx)
	}

	return loggerProvider
}

// Handler returns a slog handler from the package-level OpenTelemetry logger provider.
//
// Deprecated: use NewHandler with an explicit provider.
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

// NewProviderFromConfig creates an OpenTelemetry logger provider from explicit config.
func NewProviderFromConfig(ctx context.Context, cfg Config) (*LoggerProvider, *otelslog.HandlerOptions, error) {
	logExporter, err := otlplogs.NewExporter(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("get otel logger provider exporter: %w", err)
	}

	exp, err := otlptracegrpc.New(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("get otlptracegrpc logger: %w", err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(newResource(&cfg)),
	)

	otel.SetTracerProvider(tracerProvider)

	return &LoggerProvider{
		ctx: ctx,
		LoggerProvider: sdk.NewLoggerProvider(
			sdk.WithBatcher(logExporter),
			sdk.WithResource(newResource(&cfg)),
		),
		TracerProvider: tracerProvider,
	}, cfg.GetOptions(), nil
}

// NewProvider creates an OpenTelemetry provider from environment config.
//
// Deprecated: use NewProviderFromConfig with explicit config ownership.
func NewProvider(ctx context.Context) (*LoggerProvider, *otelslog.HandlerOptions) {
	cfg, err := GetConfigFromEnv()
	if err != nil {
		slog.Default().Error("get otel logger provider config", "err", err)
		return nil, nil
	}

	provider, opts, err := NewProviderFromConfig(ctx, *cfg)
	if err != nil {
		slog.Default().Error("create otel logger provider", "err", err)
		return nil, nil
	}

	return provider, opts
}

// NewHandler creates an OpenTelemetry slog handler from an explicit provider.
func NewHandler(provider *LoggerProvider, opts *otelslog.HandlerOptions) slog.Handler {
	return otelslog.NewOtelHandler(provider, opts)
}

func (l *LoggerProvider) Shutdown() {
	if err := l.Close(); err != nil {
		slog.Default().Error("shutdown logger provider", "err", err)
	}
}

// Close shuts down OpenTelemetry logger and tracer providers.
func (l *LoggerProvider) Close() error {
	if l == nil || l.ctx == nil {
		return nil
	}

	var errs []error

	if l.LoggerProvider != nil {
		if err := l.LoggerProvider.Shutdown(l.ctx); err != nil {
			errs = append(errs, fmt.Errorf("shutdown logger provider: %w", err))
		}
	}

	if l.TracerProvider != nil {
		if err := l.TracerProvider.Shutdown(l.ctx); err != nil {
			errs = append(errs, fmt.Errorf("shutdown tracer provider: %w", err))
		}
	}

	return errors.Join(errs...)
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
