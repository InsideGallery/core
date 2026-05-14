package statsd

import "testing"

func TestGetConfigFromEnv(t *testing.T) {
	t.Setenv("METRICS_STATSD_ADDR", "127.0.0.1:9125")
	t.Setenv("METRICS_STATSD_NAMESPACE", "custom.")

	cfg, err := getConfigFromEnv()
	if err != nil {
		t.Fatalf("getConfigFromEnv() error: %v", err)
	}

	if cfg.Addr != "127.0.0.1:9125" {
		t.Fatalf("Addr = %q", cfg.Addr)
	}

	if got := cfg.namespacePrefix(); got != "custom." {
		t.Fatalf("namespacePrefix() = %q", got)
	}
}

func TestGetConfigFromEnvUsesDefaultNamespaceWhenBlank(t *testing.T) {
	t.Setenv("METRICS_STATSD_NAMESPACE", " ")

	cfg, err := getConfigFromEnv()
	if err != nil {
		t.Fatalf("getConfigFromEnv() error: %v", err)
	}

	if got := cfg.namespacePrefix(); got != "ptolemy." {
		t.Fatalf("namespacePrefix() = %q, want ptolemy.", got)
	}
}
