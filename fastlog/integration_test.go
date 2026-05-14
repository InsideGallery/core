//go:build integration

package fastlog

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/InsideGallery/core/fastlog/handlers"
	datadoghandler "github.com/InsideGallery/core/fastlog/handlers/datadog"
	otelhandler "github.com/InsideGallery/core/fastlog/handlers/otel"
)

const (
	fastlogDatadogIntegrationEnv = "PTOLEMY_FASTLOG_DATADOG_INTEGRATION"
	fastlogOTELIntegrationEnv    = "PTOLEMY_FASTLOG_OTEL_INTEGRATION"
	fastlogIntegrationTimeout    = 5 * time.Second
)

type stoppingHandler interface {
	Stop(context.Context) error
}

func TestIntegrationDatadogHandlerExportsRecord(t *testing.T) {
	requireFastlogIntegrationSwitch(t, fastlogDatadogIntegrationEnv, "Datadog log exporter")
	requireFastlogIntegrationEnv(t, "DATADOG_API_KEY")

	handler, err := handlers.Get(datadoghandler.OutKind, handlers.FormatJSON, slog.LevelInfo)
	if err != nil {
		t.Fatalf("handlers.Get(datadog) error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), fastlogIntegrationTimeout)
	defer cancel()

	if err := handler.Handle(ctx, integrationLogRecord("datadog")); err != nil {
		t.Fatalf("Handle(datadog) error: %v", err)
	}

	stopFastlogIntegrationHandler(t, ctx, handler)
}

func TestIntegrationOTELProviderExportsRecord(t *testing.T) {
	requireFastlogIntegrationSwitch(t, fastlogOTELIntegrationEnv, "OTEL log exporter")
	requireFastlogAnyIntegrationEnv(t, "OTEL_EXPORTER_OTLP_ENDPOINT", "OTEL_EXPORTER_OTLP_LOGS_ENDPOINT")

	ctx, cancel := context.WithTimeout(context.Background(), fastlogIntegrationTimeout)
	defer cancel()

	provider, err := otelhandler.NewProvider(ctx)
	if err != nil {
		t.Fatalf("NewProvider() error: %v", err)
	}

	t.Cleanup(provider.Shutdown)

	handler, err := handlers.Get(otelhandler.OutKind, handlers.FormatJSON, slog.LevelInfo)
	if err != nil {
		t.Fatalf("handlers.Get(otel) error: %v", err)
	}

	if err := handler.Handle(ctx, integrationLogRecord("otel")); err != nil {
		t.Fatalf("Handle(otel) error: %v", err)
	}
}

func integrationLogRecord(exporter string) slog.Record {
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "ptolemy fastlog integration probe", 0)
	record.AddAttrs(
		slog.String("task", "ARCH-STD-017"),
		slog.String("exporter", exporter),
	)

	return record
}

func requireFastlogIntegrationSwitch(t *testing.T, envName, description string) {
	t.Helper()

	if strings.TrimSpace(os.Getenv(envName)) == "" {
		t.Skipf("set %s=1 to run %s integration test", envName, description)
	}
}

func requireFastlogIntegrationEnv(t *testing.T, envName string) {
	t.Helper()

	if strings.TrimSpace(os.Getenv(envName)) == "" {
		t.Fatalf("set %s for live exporter integration test", envName)
	}
}

func requireFastlogAnyIntegrationEnv(t *testing.T, envNames ...string) {
	t.Helper()

	for _, envName := range envNames {
		if strings.TrimSpace(os.Getenv(envName)) != "" {
			return
		}
	}

	t.Fatalf("set one of %s for live exporter integration test", strings.Join(envNames, ", "))
}

func stopFastlogIntegrationHandler(t *testing.T, ctx context.Context, handler slog.Handler) {
	t.Helper()

	stopper, ok := handler.(stoppingHandler)
	if !ok {
		return
	}

	if err := stopper.Stop(ctx); err != nil {
		t.Fatalf("Stop() error: %v", err)
	}
}
