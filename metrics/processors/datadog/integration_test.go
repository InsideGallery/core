//go:build integration

package datadog

import (
	"os"
	"strings"
	"testing"

	"github.com/InsideGallery/core/metrics"
)

const datadogMetricsIntegrationEnv = "PTOLEMY_METRICS_DATADOG_INTEGRATION"

func TestIntegrationProcessorExportsDogStatsDMetrics(t *testing.T) {
	requireDatadogMetricsIntegrationSwitch(t, "Datadog DogStatsD metrics exporter")
	requireDatadogMetricsAnyEnv(t, "METRICS_DATADOG_ADDR", "DD_STATSD_ADDR")

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

func requireDatadogMetricsIntegrationSwitch(t *testing.T, description string) {
	t.Helper()

	if strings.TrimSpace(os.Getenv(datadogMetricsIntegrationEnv)) == "" {
		t.Skipf("set %s=1 to run %s integration test", datadogMetricsIntegrationEnv, description)
	}
}

func requireDatadogMetricsAnyEnv(t *testing.T, envNames ...string) {
	t.Helper()

	for _, envName := range envNames {
		if strings.TrimSpace(os.Getenv(envName)) != "" {
			return
		}
	}

	t.Fatalf("set one of %s for live exporter integration test", strings.Join(envNames, ", "))
}
