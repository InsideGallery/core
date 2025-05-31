package middleware

import (
	"github.com/InsideGallery/core/queue/nats/subscriber"
)

type Middleware struct {
	chains []Chain
}

type Chain func(h subscriber.MsgHandler) subscriber.MsgHandler

func NewMiddleware(c ...Chain) *Middleware {
	return &Middleware{
		chains: c,
	}
}

func (m *Middleware) Then(next subscriber.MsgHandler) subscriber.MsgHandler {
	for i := range m.chains {
		next = m.chains[len(m.chains)-1-i](next)
	}

	return next
}

func GetChains(chains ...Chain) []Chain {
	return append([]Chain{NewRecovery().Call}, chains...)
}
