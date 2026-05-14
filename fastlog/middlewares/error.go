package middlewares

import (
	"context"
	"fmt"
	"log/slog"
)

// ErrorFormattingMiddleware converts error attributes into structured groups
// with "type" and "message" fields.
func ErrorFormattingMiddleware(
	ctx context.Context,
	record slog.Record,
	next func(context.Context, slog.Record) error,
) error {
	var attrs []slog.Attr

	record.Attrs(func(a slog.Attr) bool {
		if a.Key == "error" && a.Value.Kind() == slog.KindAny {
			if err, ok := a.Value.Any().(error); ok {
				a = slog.Group("error",
					slog.String("type", fmt.Sprintf("%T", err)),
					slog.String("message", err.Error()),
				)
			}
		}

		attrs = append(attrs, a)

		return true
	})

	newRecord := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	newRecord.AddAttrs(attrs...)

	return next(ctx, newRecord)
}
