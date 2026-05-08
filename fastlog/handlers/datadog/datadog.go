package datadog

import (
	"context"
	"errors"
	"log/slog"
	"os"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	slogdatadog "github.com/samber/slog-datadog/v2"
	slogotel "github.com/samber/slog-otel"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/InsideGallery/core/fastlog/handlers"
)

const OutKind = "datadog"

func init() {
	// Deprecated: call Setup with an explicit registry and handler factory.
	err := Setup(handlers.DefaultRegistry(), legacyHandlerFactory(context.Background()))
	if err != nil {
		slog.Default().Error("setup datadog handler", "err", err)
	}
}

// Setup registers the Datadog handler factory on an explicit registry.
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

// NewHandlerFactory creates a handler factory from explicit Datadog config.
func NewHandlerFactory(ctx context.Context, cfg Config) handlers.HandlerFactory {
	return func() slog.Handler {
		handler, err := NewHandler(ctx, cfg)
		if err != nil {
			slog.Default().Error("create datadog logger provider", "err", err)

			return nil
		}

		return handler
	}
}

func legacyHandlerFactory(ctx context.Context) handlers.HandlerFactory {
	return func() slog.Handler {
		return Handler(ctx)
	}
}

func newDatadogClient(ctx context.Context, endpoint string, apiKey string) (*datadog.APIClient, context.Context) {
	ctx = datadog.NewDefaultContext(ctx)
	ctx = context.WithValue(
		ctx,
		datadog.ContextAPIKeys,
		map[string]datadog.APIKey{"apiKeyAuth": {Key: apiKey}},
	)
	ctx = context.WithValue(
		ctx,
		datadog.ContextServerVariables,
		map[string]string{"site": endpoint},
	)
	configuration := datadog.NewConfiguration()
	apiClient := datadog.NewAPIClient(configuration)

	return apiClient, ctx
}

// NewHandler creates a Datadog slog handler from explicit config.
func NewHandler(ctx context.Context, cfg Config) (slog.Handler, error) {
	hostName, err := os.Hostname()
	if err != nil {
		slog.Default().Warn("get hostname", "err", err)
	}

	apiClient, ctx := newDatadogClient(ctx, cfg.Endpoint, cfg.APIKey)

	handler := slogdatadog.Option{
		Level:    cfg.Level,
		Client:   apiClient,
		Context:  ctx,
		Timeout:  cfg.Timeout,
		Hostname: hostName,
		Service:  cfg.Service,
		AttrFromContext: []func(ctx context.Context) []slog.Attr{
			slogotel.ExtractOtelAttrFromContext([]string{"tracing"}, "trace_id", "span_id"),
		},
	}.NewDatadogHandler()

	return handler, nil
}

// Handler creates a Datadog slog handler from environment config.
//
// Deprecated: use NewHandler with explicit config ownership.
func Handler(ctx context.Context) slog.Handler {
	cfg, err := GetConfigFromEnv()
	if err != nil {
		slog.Default().Error("get datadog logger provider config", "err", err)
		return nil
	}

	handler, err := NewHandler(ctx, *cfg)
	if err != nil {
		slog.Default().Error("create datadog logger provider", "err", err)
		return nil
	}

	return handler
}

func Tracer(
	ctx context.Context, name, spanName string, kind trace.SpanKind,
) (context.Context, trace.Span) {
	tracer := otel.Tracer(name)
	return tracer.Start(ctx, spanName, trace.WithSpanKind(kind))
}

func TracerEnd(span trace.Span) {
	span.End(trace.WithStackTrace(true))
}

func TracerWrapper(
	ctx context.Context,
	name, spanName string,
	kind trace.SpanKind,
	fn func(ctx context.Context, span trace.Span),
) {
	ctx, span := Tracer(ctx, name, spanName, kind)
	defer TracerEnd(span)

	if fn != nil {
		fn(ctx, span)
	}
}
