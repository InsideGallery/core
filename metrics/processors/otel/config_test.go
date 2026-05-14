package otel

import "testing"

func TestGetConfigFromEnv(t *testing.T) {
	t.Setenv("METRICS_OTEL_METER_NAME", "custom-meter")

	cfg, err := getConfigFromEnv()
	if err != nil {
		t.Fatalf("getConfigFromEnv() error: %v", err)
	}

	if cfg.MeterName != "custom-meter" {
		t.Fatalf("MeterName = %q", cfg.MeterName)
	}
}

func TestGetConfigFromEnvUsesDefaultWhenBlank(t *testing.T) {
	t.Setenv("METRICS_OTEL_METER_NAME", " ")

	cfg, err := getConfigFromEnv()
	if err != nil {
		t.Fatalf("getConfigFromEnv() error: %v", err)
	}

	if cfg.MeterName != defaultMeterName {
		t.Fatalf("MeterName = %q, want %q", cfg.MeterName, defaultMeterName)
	}
}
