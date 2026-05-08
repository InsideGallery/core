//go:generate mockgen -package mock -source=interface.go -destination=mock/consumer.go
package subscriber

import (
	"context"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/metric"

	"github.com/InsideGallery/core/queue/nats/client"
)

// Client is the legacy NATS SDK-shaped subscriber dependency.
//
// Deprecated: use queue/generic/subscriber/interfaces.Client for new code.
type Client interface {
	Conn() *nats.Conn
	Context() context.Context
	Logger() client.Logger
	Config() *client.Config
	QueueSubscribeSync(subject, queue string) (*nats.Subscription, error)
	Meter() metric.Meter
	WithMeter(metric.Meter)
}

// MsgHandler is the legacy NATS message handler shape.
//
// Deprecated: use queue/generic/subscriber/interfaces.MsgHandler for new code.
type MsgHandler func(ctx context.Context, msg *nats.Msg) error
