package datadog

import (
	"log/slog"
	"testing"
	"time"
)

func TestGetConfigFromEnv(t *testing.T) {
	t.Setenv("DATADOG_HOST", "localhost")
	t.Setenv("DATADOG_SERVICE", "ptolemy-test")
	t.Setenv("DATADOG_ENDPOINT", "datadoghq.com")
	t.Setenv("DATADOG_API_KEY", "test-key")
	t.Setenv("DATADOG_TIMEOUT", "250ms")
	t.Setenv("DATADOG_LEVEL", "DEBUG")

	cfg, err := getConfigFromEnv()
	if err != nil {
		t.Fatalf("getConfigFromEnv() error: %v", err)
	}

	if cfg.Host != "localhost" {
		t.Fatalf("Host = %q", cfg.Host)
	}

	if cfg.Service != "ptolemy-test" {
		t.Fatalf("Service = %q", cfg.Service)
	}

	if cfg.Endpoint != "datadoghq.com" {
		t.Fatalf("Endpoint = %q", cfg.Endpoint)
	}

	if cfg.APIKey != "test-key" {
		t.Fatalf("APIKey = %q", cfg.APIKey)
	}

	if cfg.Timeout != 250*time.Millisecond {
		t.Fatalf("Timeout = %v", cfg.Timeout)
	}

	if cfg.Level != slog.LevelDebug {
		t.Fatalf("Level = %v, want DEBUG", cfg.Level)
	}
}
