//go:build integration

package statsd

import (
	"os"
	"strings"
	"testing"

	"github.com/InsideGallery/core/metrics"
)

const statsdMetricsIntegrationEnv = "PTOLEMY_METRICS_STATSD_INTEGRATION"

func TestIntegrationProcessorExportsStatsDMetrics(t *testing.T) {
	requireStatsDMetricsIntegrationSwitch(t, "StatsD metrics exporter")
	requireStatsDMetricsEnv(t, "METRICS_STATSD_ADDR")

	rawProcessor, err := New(metrics.Config{}, "ptolemy-integration")
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	t.Cleanup(func() {
		if err := rawProcessor.Close(); err != nil {
			t.Fatalf("Close() error: %v", err)
		}
	})

	if err := rawProcessor.Count("integration.requests", 1, nil); err != nil {
		t.Fatalf("Count() error: %v", err)
	}

	if err := rawProcessor.Gauge("integration.active", 1, nil); err != nil {
		t.Fatalf("Gauge() error: %v", err)
	}

	if err := rawProcessor.Distribution("integration.latency", 1.25, nil); err != nil {
		t.Fatalf("Distribution() error: %v", err)
	}
}

func requireStatsDMetricsIntegrationSwitch(t *testing.T, description string) {
	t.Helper()

	if strings.TrimSpace(os.Getenv(statsdMetricsIntegrationEnv)) == "" {
		t.Skipf("set %s=1 to run %s integration test", statsdMetricsIntegrationEnv, description)
	}
}

func requireStatsDMetricsEnv(t *testing.T, envName string) {
	t.Helper()

	if strings.TrimSpace(os.Getenv(envName)) == "" {
		t.Fatalf("set %s for live exporter integration test", envName)
	}
}
