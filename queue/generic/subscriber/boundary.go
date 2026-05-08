package subscriber

import "github.com/InsideGallery/core/queue/generic/subscriber/interfaces"

// Client is the core-owned generic subscriber dependency contract.
type Client = interfaces.Client

// Message is the core-owned generic subscriber message contract.
type Message = interfaces.Msg

// MessageHandler handles generic subscriber messages.
type MessageHandler = interfaces.MsgHandler

// SubscriptionHandle is the core-owned generic subscription contract.
type SubscriptionHandle = interfaces.Subscription

// NewMessageSubscriber creates a generic subscriber without importing the interfaces subpackage.
func NewMessageSubscriber(client Client) *Subscriber {
	return NewSubscriber(client)
}
