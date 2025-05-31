package app

import (
	"context"
	"testing"

	_ "github.com/InsideGallery/core/fastlog/handlers/stderr"

	"github.com/InsideGallery/core/fastlog/metrics"
	"github.com/InsideGallery/core/queue/nats/middleware"
	"github.com/InsideGallery/core/queue/nats/subscriber"
)

func TestWeb(t *testing.T) {
	t.Skip() // test for validate how apps works
	ctx := context.Background()
	WebMain(ctx, ":8090", "Test Server", nil)
}

func TestNATS(t *testing.T) {
	t.Skip() // test for validate how apps works
	ctx := context.Background()
	NATSMain(ctx, func(_ context.Context, _ *metrics.OTLPMetric, _ *middleware.Middleware, _ *subscriber.Subscriber) error {
		panic("test")
	})
}
