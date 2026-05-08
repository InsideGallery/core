package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"

	corenats "github.com/InsideGallery/core/queue/nats"
)

// MessageChain composes core-owned NATS message handlers.
type MessageChain func(h corenats.MessageHandler) corenats.MessageHandler

// MessageMiddleware contains core-owned middleware chains for NATS messages.
type MessageMiddleware struct {
	chains []MessageChain
}

// NewMessageMiddleware creates middleware without exposing NATS SDK message types.
func NewMessageMiddleware(chains ...MessageChain) *MessageMiddleware {
	return &MessageMiddleware{
		chains: chains,
	}
}

// AddChain appends a core-owned message middleware chain.
func (m *MessageMiddleware) AddChain(chain MessageChain) {
	m.chains = append(m.chains, chain)
}

// Then wraps the handler with all configured message chains.
func (m *MessageMiddleware) Then(next corenats.MessageHandler) corenats.MessageHandler {
	for i := range m.chains {
		next = m.chains[len(m.chains)-1-i](next)
	}

	return next
}

// Copy returns a copy of the middleware chain.
func (m *MessageMiddleware) Copy() *MessageMiddleware {
	return NewMessageMiddleware(m.chains...)
}

// Merge appends chains from other message middleware values.
func (m *MessageMiddleware) Merge(middlewares ...*MessageMiddleware) {
	for _, middleware := range middlewares {
		for _, chain := range middleware.chains {
			m.AddChain(chain)
		}
	}
}

// MessageRecovery recovers panics in core-owned message handlers.
type MessageRecovery struct{}

// NewMessageRecovery creates panic recovery middleware for core-owned messages.
func NewMessageRecovery() *MessageRecovery {
	return &MessageRecovery{}
}

// Call wraps a core-owned message handler with panic recovery.
func (r *MessageRecovery) Call(next corenats.MessageHandler) corenats.MessageHandler {
	return func(ctx context.Context, msg corenats.Message) error {
		defer func() {
			if value := recover(); value != nil {
				slog.Default().Error("panic recovered",
					"error", fmt.Sprintf("%v", value),
					"stack", string(debug.Stack()),
				)
			}
		}()

		return next(ctx, msg)
	}
}

// GetMessageChains prepends the default recovery chain to core-owned message middleware.
func GetMessageChains(chains ...MessageChain) []MessageChain {
	return append([]MessageChain{NewMessageRecovery().Call}, chains...)
}
