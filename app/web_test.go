package app

import (
	"context"
	"testing"

	"github.com/InsideGallery/core/metrics"
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
	NATSMain(ctx, func(_ context.Context, _ *metrics.Client, _ *middleware.Middleware, _ *subscriber.Subscriber) error {
		panic("test")
	})
}
