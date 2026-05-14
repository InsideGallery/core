package middlewares

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"runtime"
)

const callerDepth = 4

// CallerMiddleware adds source file and line information to log records.
func CallerMiddleware(ctx context.Context, record slog.Record, next func(context.Context, slog.Record) error) error {
	var attrs []slog.Attr

	record.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	attrs = append(attrs, slog.String("caller", caller(callerDepth)))

	newRecord := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	newRecord.AddAttrs(attrs...)

	return next(ctx, newRecord)
}

func caller(depth int) string {
	_, file, line, ok := runtime.Caller(depth + 1)
	if !ok {
		return "unknown"
	}

	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}
