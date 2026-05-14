package middlewares

import (
	"context"
	"log/slog"

	"github.com/FrogoAI/set"
	slogmulti "github.com/samber/slog-multi"
)

const maskValue = "*******"

var maskKeySet = set.NewGenericDataSet[string]( //nolint:gochecknoglobals // immutable lookup table
	"password",
	"email",
	"phone",
)

// NewGDPRMiddleware returns a slog middleware that masks common PII fields.
func NewGDPRMiddleware() slogmulti.Middleware {
	return func(next slog.Handler) slog.Handler {
		return &gdprMiddleware{
			next: next,
		}
	}
}

type gdprMiddleware struct {
	next      slog.Handler
	anonymize bool
}

func (h *gdprMiddleware) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *gdprMiddleware) Handle(ctx context.Context, record slog.Record) error {
	var attrs []slog.Attr

	record.Attrs(func(attr slog.Attr) bool {
		if mightContainPII(attr.Key) {
			attrs = append(attrs, anonymize(attr))

			return true
		}

		attrs = append(attrs, attr)

		return true
	})

	record = slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	record.AddAttrs(attrs...)

	return h.next.Handle(ctx, record)
}

func (h *gdprMiddleware) WithAttrs(attrs []slog.Attr) slog.Handler {
	for i := range attrs {
		if h.anonymize || mightContainPII(attrs[i].Key) {
			attrs[i] = anonymize(attrs[i])
		}
	}

	return &gdprMiddleware{
		next:      h.next.WithAttrs(attrs),
		anonymize: h.anonymize,
	}
}

func (h *gdprMiddleware) WithGroup(name string) slog.Handler {
	return &gdprMiddleware{
		next:      h.next.WithGroup(name),
		anonymize: h.anonymize || mightContainPII(name),
	}
}

func mightContainPII(key string) bool {
	return maskKeySet.Contains(key)
}

func anonymize(attr slog.Attr) slog.Attr {
	if attr.Value.Kind() != slog.KindGroup {
		return slog.String(attr.Key, maskValue)
	}

	attrs := attr.Value.Group()
	for i := range attrs {
		attrs[i] = anonymize(attrs[i])
	}

	args := make([]any, len(attrs))
	for i := range attrs {
		args[i] = attrs[i]
	}

	return slog.Group(attr.Key, args...)
}
