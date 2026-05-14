package fastlog

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/InsideGallery/core/fastlog/handlers"
)

func TestGetConfigFromEnvParsesLoggingEnv(t *testing.T) {
	t.Setenv("LOG_OUTPUTS", "nop:json,nop:text")
	t.Setenv("LOG_LEVEL", "DEBUG")
	t.Setenv("LOG_CALLER", "false")
	t.Setenv("LOG_ERROR_FORMATTING", "true")

	cfg, err := GetConfigFromEnv()
	if err != nil {
		t.Fatalf("GetConfigFromEnv() error: %v", err)
	}

	wantOutputs := []string{"nop:json", "nop:text"}
	if len(cfg.Outputs) != len(wantOutputs) {
		t.Fatalf("Outputs = %v, want %v", cfg.Outputs, wantOutputs)
	}

	for i := range wantOutputs {
		if cfg.Outputs[i] != wantOutputs[i] {
			t.Fatalf("Outputs = %v, want %v", cfg.Outputs, wantOutputs)
		}
	}

	if cfg.Level != slog.LevelDebug {
		t.Fatalf("Level = %v, want DEBUG", cfg.Level)
	}

	if cfg.Caller {
		t.Fatal("Caller = true, want false")
	}

	if !cfg.ErrorFormatting {
		t.Fatal("ErrorFormatting = false, want true")
	}
}

func TestGetConfigFromEnvReturnsParseError(t *testing.T) {
	t.Setenv("LOG_LEVEL", "not-a-level")

	if _, err := GetConfigFromEnv(); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestConfigGetHandlerFallsBackToNopForUnknownOutput(t *testing.T) {
	cfg := Config{
		Outputs: []string{"missing:json"},
		Level:   slog.LevelInfo,
	}

	handler, err := cfg.GetHandler()
	if !errors.Is(err, handlers.ErrNotFoundHandler) {
		t.Fatalf("GetHandler() error = %v, want ErrNotFoundHandler", err)
	}

	if handler == nil {
		t.Fatal("expected fallback handler")
	}

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "fallback", 0)
	if err := handler.Handle(context.Background(), record); err != nil {
		t.Fatalf("Handle() error: %v", err)
	}
}

func TestConfigGetHandlerSkipsMalformedOutputAndFallsBack(t *testing.T) {
	cfg := Config{
		Outputs: []string{"malformed"},
		Level:   slog.LevelInfo,
	}

	handler, err := cfg.GetHandler()
	if err != nil {
		t.Fatalf("GetHandler() error: %v", err)
	}

	if handler == nil {
		t.Fatal("expected fallback handler")
	}
}

func TestSetupDefaultLoggerSetsDefaultLogger(t *testing.T) {
	previous := slog.Default()

	t.Cleanup(func() {
		slog.SetDefault(previous)
	})

	cfg := &Config{
		Outputs: []string{"nop:json"},
		Level:   slog.LevelInfo,
		Caller:  false,
	}

	if err := SetupDefaultLogger(cfg); err != nil {
		t.Fatalf("SetupDefaultLogger() error: %v", err)
	}

	if slog.Default() == previous {
		t.Fatal("expected default logger to be replaced")
	}

	if !slog.Default().Handler().Enabled(context.Background(), slog.LevelInfo) {
		t.Fatal("default logger should enable info logs")
	}
}

func TestSetupDefaultLoggerReturnsErrors(t *testing.T) {
	if err := SetupDefaultLogger(nil); err == nil {
		t.Fatal("expected setup error")
	}
}
