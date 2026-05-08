package webserver

import "github.com/gofiber/fiber/v3"

// Middleware contains methods to call before handle request
type Middleware struct {
	chains []Chain
}

// RouteChain composes core-owned route handlers.
type RouteChain func(h RouteHandler) RouteHandler

// RouteMiddleware contains core-owned route middleware chains.
type RouteMiddleware struct {
	chains []RouteChain
}

// Chain single sequence
//
// Deprecated: use core-owned Client/Runtime contracts for new code that does not need Fiber middleware.
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

// NewRouteMiddleware returns middleware for core-owned route handlers.
func NewRouteMiddleware(chains ...RouteChain) *RouteMiddleware {
	return &RouteMiddleware{
		chains: chains,
	}
}

// AddChain appends a core-owned route middleware chain.
func (m *RouteMiddleware) AddChain(chain RouteChain) {
	m.chains = append(m.chains, chain)
}

// Then wraps the route handler with all configured route chains.
func (m *RouteMiddleware) Then(handler RouteHandler) RouteHandler {
	for i := range m.chains {
		handler = m.chains[len(m.chains)-1-i](handler)
	}

	return handler
}

// Copy returns a copy of the route middleware chain.
func (m *RouteMiddleware) Copy() *RouteMiddleware {
	return NewRouteMiddleware(m.chains...)
}

// Merge appends chains from other route middleware values.
func (m *RouteMiddleware) Merge(middlewares ...*RouteMiddleware) {
	for _, middleware := range middlewares {
		for _, chain := range middleware.chains {
			m.AddChain(chain)
		}
	}
}
