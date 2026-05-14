package otel

import (
	"log/slog"
	"testing"
)

func TestGetConfigFromEnv(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "ptolemy-test")
	t.Setenv("OTEL_SERVICE_VERSION", "v2.0.0")
	t.Setenv("OTEL_NAMESPACE", "observability")
	t.Setenv("OTEL_LEVEL", "DEBUG")

	cfg, err := getConfigFromEnv()
	if err != nil {
		t.Fatalf("getConfigFromEnv() error: %v", err)
	}

	if cfg.ServiceName != "ptolemy-test" {
		t.Fatalf("ServiceName = %q", cfg.ServiceName)
	}

	if cfg.ServiceVersion != "v2.0.0" {
		t.Fatalf("ServiceVersion = %q", cfg.ServiceVersion)
	}

	if cfg.Namespace != "observability" {
		t.Fatalf("Namespace = %q", cfg.Namespace)
	}

	if cfg.Level != slog.LevelDebug {
		t.Fatalf("Level = %v, want DEBUG", cfg.Level)
	}
}

func TestNewHandlerUsesConfiguredService(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "ptolemy-test")

	handler, err := newHandler()
	if err != nil {
		t.Fatalf("newHandler() error: %v", err)
	}

	if handler == nil {
		t.Fatal("expected handler")
	}
}
