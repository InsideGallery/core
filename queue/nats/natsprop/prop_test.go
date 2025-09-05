package natsprop

import (
	"context"
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// use trace propagation
// due to NewCompositeTextMapPropagator it's possible send both of them
func TestTrace(t *testing.T) {
	prop := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{})

	ctx := context.Background()
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{0x03},
		SpanID:  trace.SpanID{0x03},
	})
	ctx = trace.ContextWithRemoteSpanContext(ctx, sc)

	msg := new(nats.Msg)

	t.Run("inject", func(t *testing.T) {
		Inject(ctx, msg, WithPropagators(prop))
		assert.Len(t, msg.Header, 1)
	})

	// depend on inject
	t.Run("extract", func(t *testing.T) {
		spanContext := Extract(ctx, msg, WithPropagators(prop))
		assert.Equal(t, sc.SpanID(), spanContext.SpanID())
		assert.Equal(t, sc.TraceID(), spanContext.TraceID())
	})
}
