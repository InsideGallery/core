package datadog

import (
	"context"
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
	handlers.RegisterHandler(OutKind,
		Handler(context.Background()),
	)
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

func Handler(ctx context.Context) slog.Handler {
	cfg, err := GetConfigFromEnv()
	if err != nil {
		slog.Default().Error("Error get datadog logger provider config", "err", err)
		return nil
	}

	hostName, err := os.Hostname()
	if err != nil {
		slog.Default().Warn("Error get hostname", "err", err)
	}

	apiClient, ctx := newDatadogClient(ctx, cfg.Endpoint, cfg.APIKey)

	return slogdatadog.Option{
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
