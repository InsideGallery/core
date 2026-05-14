package app

import "testing"

func TestAppMetricsConfigUsesConfiguredProcessors(t *testing.T) {
	t.Setenv("METRICS_PROCESSORS", "datadog,statsd")

	cfg, err := appMetricsConfig()
	if err != nil {
		t.Fatalf("appMetricsConfig() error: %v", err)
	}

	got := cfg.EnabledProcessors()
	want := []string{"datadog", "statsd"}

	if len(got) != len(want) {
		t.Fatalf("processors = %v, want %v", got, want)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("processors = %v, want %v", got, want)
		}
	}
}

func TestAppMetricsConfigKeepsMetricsDisabled(t *testing.T) {
	t.Setenv("METRICS_PROCESSORS", "none")

	cfg, err := appMetricsConfig()
	if err != nil {
		t.Fatalf("appMetricsConfig() error: %v", err)
	}

	if cfg.Enabled() {
		t.Fatal("expected disabled metrics")
	}
}
