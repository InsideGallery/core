package middlewares

import (
	"context"
	"log/slog"

	"github.com/samber/lo"
	slogmulti "github.com/samber/slog-multi"

	"github.com/InsideGallery/core/memory/set"
)

var maskKeySet = set.NewGenericDataSet[string](
	"password",
	"email",
	"phone",
)

const maskValue = "*******"

func NewGDPRMiddleware() slogmulti.Middleware {
	return func(next slog.Handler) slog.Handler {
		return &gdprMiddleware{
			next:      next,
			anonymize: false,
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
		} else {
			attrs = append(attrs, attr)
		}

		return true
	})

	// new record with anonymized data
	record = slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	record.AddAttrs(attrs...)

	return h.next.Handle(ctx, record)
}

func (h *gdprMiddleware) WithAttrs(attrs []slog.Attr) slog.Handler {
	if h.anonymize {
		for i := range attrs {
			attrs[i] = anonymize(attrs[i])
		}
	}

	for i := range attrs {
		if mightContainPII(attrs[i].Key) {
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
	k := attr.Key
	v := attr.Value
	kind := attr.Value.Kind()

	switch kind {
	case slog.KindGroup:
		attrs := v.Group()
		for i := range attrs {
			attrs[i] = anonymize(attrs[i])
		}

		return slog.Group(k, lo.ToAnySlice(attrs)...)
	default:
		return slog.String(k, maskValue)
	}
}
