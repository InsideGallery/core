# fastlog/handlers/otel

Import path: `github.com/InsideGallery/core/fastlog/handlers/otel`

This package registers the `otel` fastlog output and provides an OpenTelemetry log and trace provider. The handler uses
OTLP log export and also configures a global OpenTelemetry tracer provider.

## Main APIs

- `OutKind` is the registry key: `otel`.
- `LoggerProvider` owns the log and trace providers.
- `Default(ctx context.Context)` returns the singleton provider, initializing it once.
- `NewProvider(ctx context.Context)` creates providers from environment-backed config.
- `(*LoggerProvider).Shutdown()` shuts down both providers and logs shutdown errors.
- `(*LoggerProvider).Tracer`, `TracerEnd`, and `TracerWrapper` help create and finish spans.

## Usage

```go
package main

import (
	_ "github.com/InsideGallery/core/fastlog/handlers/otel"

	"github.com/InsideGallery/core/fastlog"
)

func configureLogging() error {
	cfg, err := fastlog.GetConfigFromEnv()
	if err != nil {
		return err
	}

	return fastlog.SetupDefaultLogger(cfg)
}
```

Run with `LOG_OUTPUTS=otel:json`. The registry format field is ignored because this package registers a complete
handler factory.

## Configuration

The package reads the `OTEL` prefix:

- `OTEL_SERVICE_NAME`: service name resource attribute, default `ptolemy`.
- `OTEL_SERVICE_VERSION`: service version resource attribute, default `v1.0.0`.
- `OTEL_NAMESPACE`: service namespace resource attribute, default `default`.
- `OTEL_LEVEL`: handler level, default `INFO`.

The underlying OpenTelemetry exporters use standard OTLP exporter configuration. The integration test checks for either
`OTEL_EXPORTER_OTLP_ENDPOINT` or `OTEL_EXPORTER_OTLP_LOGS_ENDPOINT`, gated by
`PTOLEMY_FASTLOG_OTEL_INTEGRATION=1`.
