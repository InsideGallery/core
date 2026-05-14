package datadog

import (
	"testing"

	"github.com/InsideGallery/core/metrics"
)

func TestGetConfigFromEnv(t *testing.T) {
	t.Setenv("METRICS_DATADOG_ADDR", "127.0.0.1:8125")
	t.Setenv("METRICS_DATADOG_NAMESPACE", "custom")

	cfg, err := getConfigFromEnv()
	if err != nil {
		t.Fatalf("getConfigFromEnv() error: %v", err)
	}

	if cfg.Addr != "127.0.0.1:8125" {
		t.Fatalf("Addr = %q", cfg.Addr)
	}

	if got := cfg.namespacePrefix(); got != "custom." {
		t.Fatalf("namespacePrefix() = %q", got)
	}
}

func TestGetConfigFromEnvLegacyAddr(t *testing.T) {
	t.Setenv("DD_STATSD_ADDR", "127.0.0.1:8125")

	cfg, err := getConfigFromEnv()
	if err != nil {
		t.Fatalf("getConfigFromEnv() error: %v", err)
	}

	if cfg.Addr != "127.0.0.1:8125" {
		t.Fatalf("Addr = %q", cfg.Addr)
	}
}

func TestGetConfigFromEnvUsesDefaultNamespaceWhenBlank(t *testing.T) {
	t.Setenv("METRICS_DATADOG_NAMESPACE", " ")

	cfg, err := getConfigFromEnv()
	if err != nil {
		t.Fatalf("getConfigFromEnv() error: %v", err)
	}

	if got := cfg.namespacePrefix(); got != "ptolemy." {
		t.Fatalf("namespacePrefix() = %q, want ptolemy.", got)
	}
}

func TestNewRequiresAddress(t *testing.T) {
	t.Setenv("METRICS_DATADOG_ADDR", "")
	t.Setenv("DD_STATSD_ADDR", "")

	if _, err := New(metrics.Config{}, "test-svc"); err == nil {
		t.Fatal("expected missing address error")
	}
}
