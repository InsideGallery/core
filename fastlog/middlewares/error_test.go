package middlewares

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestErrorFormattingMiddlewareConvertsErrorAttribute(t *testing.T) {
	record := slog.NewRecord(time.Now(), slog.LevelError, "failed", 0)
	record.AddAttrs(
		slog.Any("error", errors.New("database unavailable")),
		slog.String("component", "repository"),
	)

	var captured slog.Record

	err := ErrorFormattingMiddleware(context.Background(), record, func(_ context.Context, nextRecord slog.Record) error {
		captured = nextRecord

		return nil
	})
	if err != nil {
		t.Fatalf("ErrorFormattingMiddleware() error: %v", err)
	}

	attrs := recordAttrs(captured)
	if attrs["component"].String() != "repository" {
		t.Fatalf("component = %q, want repository", attrs["component"].String())
	}

	errorAttrs := groupAttrs(attrs["error"])
	if errorAttrs["message"].String() != "database unavailable" {
		t.Fatalf("error.message = %q", errorAttrs["message"].String())
	}

	if !strings.Contains(errorAttrs["type"].String(), "errorString") {
		t.Fatalf("error.type = %q, want errorString", errorAttrs["type"].String())
	}
}

func TestErrorFormattingMiddlewareLeavesNonErrorAttribute(t *testing.T) {
	record := slog.NewRecord(time.Now(), slog.LevelError, "failed", 0)
	record.AddAttrs(slog.String("error", "plain text"))

	var captured slog.Record

	err := ErrorFormattingMiddleware(context.Background(), record, func(_ context.Context, nextRecord slog.Record) error {
		captured = nextRecord

		return nil
	})
	if err != nil {
		t.Fatalf("ErrorFormattingMiddleware() error: %v", err)
	}

	attrs := recordAttrs(captured)
	if attrs["error"].String() != "plain text" {
		t.Fatalf("error = %q, want plain text", attrs["error"].String())
	}
}

func groupAttrs(value slog.Value) map[string]slog.Value {
	attrs := make(map[string]slog.Value)

	for _, attr := range value.Group() {
		attrs[attr.Key] = attr.Value
	}

	return attrs
}
