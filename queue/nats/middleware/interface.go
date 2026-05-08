package middleware

import (
	"github.com/InsideGallery/core/queue/nats/subscriber"
)

// Middleware contains legacy NATS SDK-shaped middleware chains.
//
// Legacy: use MessageMiddleware for new code.
type Middleware struct {
	chains []Chain
}

// Chain composes legacy NATS SDK-shaped handlers.
//
// Legacy: use MessageChain for new code.
//
//nolint:staticcheck // legacy middleware keeps NATS handler shim
type Chain func(h subscriber.MsgHandler) subscriber.MsgHandler

// NewMiddleware creates legacy NATS SDK-shaped middleware.
//
// Legacy: use NewMessageMiddleware for new code.
func NewMiddleware(c ...Chain) *Middleware {
	return &Middleware{
		chains: c,
	}
}

// Then wraps a legacy NATS SDK-shaped handler.
//
// Legacy: use MessageMiddleware.Then for new code.
//
//nolint:staticcheck // legacy middleware keeps NATS handler shim
func (m *Middleware) Then(next subscriber.MsgHandler) subscriber.MsgHandler {
	for i := range m.chains {
		next = m.chains[len(m.chains)-1-i](next)
	}

	return next
}

// GetChains prepends legacy recovery middleware.
//
// Legacy: use GetMessageChains for new code.
func GetChains(chains ...Chain) []Chain {
	return append([]Chain{NewRecovery().Call}, chains...)
}
