//go:generate mockgen -package mock -source=interface.go -destination=mock/publisher.go
package publisher

import "github.com/nats-io/nats.go"

// Client is the legacy NATS SDK-shaped publisher dependency.
//
// Deprecated: use github.com/InsideGallery/core/queue/nats.Publisher for new code.
type Client interface {
	Conn() *nats.Conn
}
