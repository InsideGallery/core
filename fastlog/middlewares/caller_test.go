package middlewares

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestCallerMiddlewareAddsCallerAndPreservesAttributes(t *testing.T) {
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "hello", 0)
	record.AddAttrs(slog.String("request_id", "req-1"))

	var captured slog.Record

	err := CallerMiddleware(context.Background(), record, func(_ context.Context, nextRecord slog.Record) error {
		captured = nextRecord

		return nil
	})
	if err != nil {
		t.Fatalf("CallerMiddleware() error: %v", err)
	}

	attrs := recordAttrs(captured)
	if attrs["request_id"].String() != "req-1" {
		t.Fatalf("request_id = %q, want req-1", attrs["request_id"].String())
	}

	caller := attrs["caller"].String()
	if caller == "" {
		t.Fatal("expected caller attribute")
	}
}

func TestCallerReturnsFileAndLine(t *testing.T) {
	got := caller(0)
	if got == "" || got == "unknown" || !strings.Contains(got, ":") {
		t.Fatalf("caller(0) = %q, want file:line", got)
	}
}

func recordAttrs(record slog.Record) map[string]slog.Value {
	attrs := make(map[string]slog.Value)

	record.Attrs(func(attr slog.Attr) bool {
		attrs[attr.Key] = attr.Value

		return true
	})

	return attrs
}
