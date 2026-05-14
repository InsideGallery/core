package middlewares

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"
)

var (
	benchmarkMiddlewareAttrCount int
	benchmarkMiddlewareErr       error
)

func BenchmarkCallerMiddleware(b *testing.B) {
	ctx := context.Background()
	record := slog.NewRecord(time.Unix(0, 0), slog.LevelInfo, "request complete", 0)
	record.AddAttrs(
		slog.String("route", "/v2/notifyapi/notifications"),
		slog.Int("status", 200),
	)
	next := func(_ context.Context, record slog.Record) error {
		benchmarkMiddlewareAttrCount = record.NumAttrs()

		return nil
	}

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if err := CallerMiddleware(ctx, record, next); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkErrorFormattingMiddleware(b *testing.B) {
	ctx := context.Background()
	err := errors.New("publish failed")
	record := slog.NewRecord(time.Unix(0, 0), slog.LevelError, "notification publish failed", 0)
	record.AddAttrs(
		slog.String("subject", "ptolemy.notify.email"),
		slog.Any("error", err),
	)
	next := func(_ context.Context, record slog.Record) error {
		benchmarkMiddlewareAttrCount = record.NumAttrs()

		return nil
	}

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		benchmarkMiddlewareErr = ErrorFormattingMiddleware(ctx, record, next)
		if benchmarkMiddlewareErr != nil {
			b.Fatal(benchmarkMiddlewareErr)
		}
	}
}
