package metrics //nolint:revive // package name matches directory; runtime/metrics is a sub-package

import "testing"

func TestConfig_Enabled(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want bool
	}{
		{"empty config is disabled", Config{}, false},
		{"processor list is enabled", Config{Processors: []string{"prometheus"}}, true},
		{"none disables processors", Config{Processors: []string{"none"}}, false},
		{"off disables processors", Config{Processors: []string{"off"}}, false},
		{"disabled disables processors", Config{Processors: []string{"disabled"}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.Enabled(); got != tt.want {
				t.Errorf("Config.Enabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_EnabledProcessors(t *testing.T) {
	cfg := Config{Processors: []string{"prometheus,datadog", "statsd", "prometheus", "none"}}

	got := cfg.EnabledProcessors()
	want := []string{"prometheus", "datadog", "statsd"}

	assertProcessors(t, got, want)
}

func TestGetEnvConfig_DefaultPrefix(t *testing.T) {
	t.Setenv("METRICS_PROCESSORS", "prometheus,statsd")

	cfg, err := GetEnvConfig()
	if err != nil {
		t.Fatalf("GetEnvConfig() error: %v", err)
	}

	assertProcessors(t, cfg.EnabledProcessors(), []string{"prometheus", "statsd"})
}

func TestGetEnvConfig_CustomPrefix(t *testing.T) {
	t.Setenv("CUSTOM_PROCESSORS", "datadog")

	cfg, err := GetEnvConfig("CUSTOM")
	if err != nil {
		t.Fatalf("GetEnvConfig(CUSTOM) error: %v", err)
	}

	assertProcessors(t, cfg.EnabledProcessors(), []string{"datadog"})
}

func TestGetEnvConfig_UnsetDefaultsToPrometheus(t *testing.T) {
	cfg, err := GetEnvConfig()
	if err != nil {
		t.Fatalf("GetEnvConfig() error: %v", err)
	}

	assertProcessors(t, cfg.EnabledProcessors(), []string{"prometheus"})
}

func TestGetEnvConfig_Disabled(t *testing.T) {
	t.Setenv("METRICS_PROCESSORS", "none")

	cfg, err := GetEnvConfig()
	if err != nil {
		t.Fatalf("GetEnvConfig() error: %v", err)
	}

	if cfg.Enabled() {
		t.Fatal("expected disabled metrics")
	}
}

func assertProcessors(t *testing.T, got, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("processors = %v, want %v", got, want)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("processors = %v, want %v", got, want)
		}
	}
}
