package webserver

import "github.com/gofiber/fiber/v2"

// Middleware contains methods to call before handle request
type Middleware struct {
	chains []Chain
}

// Chain single sequence
type Chain func(h fiber.Handler) fiber.Handler

// NewMiddleware return new natsmiddleware
func NewMiddleware(c ...Chain) *Middleware {
	return &Middleware{
		chains: c,
	}
}

// AddChain add natsmiddleware to execute
func (m *Middleware) AddChain(c Chain) {
	m.chains = append(m.chains, c)
}

// Then return router handler
func (m *Middleware) Then(h fiber.Handler) fiber.Handler {
	for i := range m.chains {
		h = m.chains[len(m.chains)-1-i](h)
	}

	return h
}

// Copy return copied sequence
func (m *Middleware) Copy() *Middleware {
	return NewMiddleware(m.chains...)
}

// Merge merge logmiddlewares into current
func (m *Middleware) Merge(middlewares ...*Middleware) {
	for _, middleware := range middlewares {
		for _, chain := range middleware.chains {
			m.AddChain(chain)
		}
	}
}
