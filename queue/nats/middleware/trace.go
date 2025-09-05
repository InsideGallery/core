package middleware

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/InsideGallery/core/fastlog/handlers/otel"
	"github.com/InsideGallery/core/queue/nats/natsprop"
	"github.com/InsideGallery/core/queue/nats/subscriber"
)

const (
	TracerName = "nats"
)

// Tracer for subscribers implementing Middleware
type Tracer struct{}

func NewTracer() *Tracer {
	return &Tracer{}
}

func (t *Tracer) Call(next subscriber.MsgHandler) subscriber.MsgHandler {
	return func(ctx context.Context, msg *nats.Msg) error {
		defer func() {
			if rval := recover(); rval != nil {
				slog.Default().Error("NATS Tracer", "rval", rval)
				return
			}
		}()

		opr := defaultOperationFn(msg)
		spanContext := natsprop.Extract(ctx, msg)
		ctx = trace.ContextWithRemoteSpanContext(ctx, spanContext)

		var err error

		otel.Default(ctx).TracerWrapper(ctx, TracerName, opr, trace.SpanKindConsumer,
			func(ctx context.Context, span trace.Span) {
				natsprop.Inject(ctx, msg)

				if next == nil {
					span.SetStatus(codes.Error, "No handler set")
					span.RecordError(err)

					return
				}

				err = next(ctx, msg)
				if err != nil {
					span.SetStatus(codes.Error, err.Error())
					span.RecordError(err)
				} else {
					span.SetStatus(codes.Ok, "")
				}
			},
		)

		return err
	}
}

func defaultOperationFn(msg *nats.Msg) string {
	if msg.Sub != nil {
		return fmt.Sprintf("NATS:%s/%s", msg.Sub.Queue, msg.Subject)
	}

	return fmt.Sprintf("NATS:%s", msg.Subject)
}
