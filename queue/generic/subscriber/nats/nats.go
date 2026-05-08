// Package nats provides NATS adapters for the generic subscriber package.
//
// New code should import this focused adapter path:
//
//	import "github.com/InsideGallery/core/queue/generic/subscriber/nats"
//
//	subscriber := nats.NewSubscriber(conn)
//
// Compatibility: queue/generic/subscriber/driver remains available for existing
// consumers. Prefer this package for NATS-specific adapter aliases so the generic
// subscriber root package does not grow provider-specific names.
package nats

import (
	legacy "github.com/InsideGallery/core/queue/generic/subscriber/driver"
	"github.com/InsideGallery/core/queue/nats/client"
)

// Message adapts a NATS message to the generic subscriber message contract.
type Message = legacy.NATSMsg

// Subscription adapts a NATS subscription to the generic subscriber subscription contract.
type Subscription = legacy.NATSSubscription

// Config adapts NATS client configuration to the generic subscriber config contract.
type Config = legacy.NATSConfig

// Subscriber adapts a NATS client to the generic subscriber client contract.
type Subscriber = legacy.NATSSubscriber

// NewSubscriber returns a NATS-backed generic subscriber client.
func NewSubscriber(conn *client.Client) *Subscriber {
	return legacy.NewNATSSubscriber(conn)
}
