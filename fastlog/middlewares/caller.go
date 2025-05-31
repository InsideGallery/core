package middlewares

import (
	"context"
	"fmt"
	"log/slog"
	"path"
	"runtime"
)

const depth = 4

func CallerMiddleware(ctx context.Context, record slog.Record, next func(context.Context, slog.Record) error) error {
	var attrs []slog.Attr

	record.Attrs(func(attr slog.Attr) bool {
		attrs = append(attrs, attr)
		return true
	})
	attrs = append(attrs, slog.String("caller", Caller(depth)))

	// new record with formatted error
	record = slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	record.AddAttrs(attrs...)

	return next(ctx, record)
}

func Caller(depth int) (fileAndLine string) {
	if _, file, line, ok := runtime.Caller(depth + 1); ok {
		fileAndLine = fmt.Sprintf("%s:%d", path.Base(file), line)
	}

	return
}
