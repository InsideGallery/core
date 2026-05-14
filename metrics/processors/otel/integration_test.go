//go:build integration

package otel

import (
	"os"
	"strings"
	"testing"

	"github.com/InsideGallery/core/metrics"
)

const otelMetricsIntegrationEnv = "PTOLEMY_METRICS_OTEL_INTEGRATION"

func TestIntegrationProcessorRecordsToConfiguredMeterProvider(t *testing.T) {
	requireOTELMetricsIntegrationSwitch(t, "OTEL metrics exporter")

	rawProcessor, err := New(metrics.Config{}, "ptolemy-integration")
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	t.Cleanup(func() {
		if err := rawProcessor.Close(); err != nil {
			t.Fatalf("Close() error: %v", err)
		}
	})

	if err := rawProcessor.Count("integration.requests", 1, []string{"task:ARCH-STD-017"}); err != nil {
		t.Fatalf("Count() error: %v", err)
	}

	if err := rawProcessor.Gauge("integration.active", 1, []string{"task:ARCH-STD-017"}); err != nil {
		t.Fatalf("Gauge() error: %v", err)
	}

	if err := rawProcessor.Distribution("integration.latency", 1.25, []string{"task:ARCH-STD-017"}); err != nil {
		t.Fatalf("Distribution() error: %v", err)
	}
}

func requireOTELMetricsIntegrationSwitch(t *testing.T, description string) {
	t.Helper()

	if strings.TrimSpace(os.Getenv(otelMetricsIntegrationEnv)) == "" {
		t.Skipf("set %s=1 to run %s integration test", otelMetricsIntegrationEnv, description)
	}
}
