//go:generate mockgen -package mock -source=interface.go -destination=mock/publisher.go
package publisher

import "github.com/nats-io/nats.go"

type Client interface {
	Conn() *nats.Conn
}
