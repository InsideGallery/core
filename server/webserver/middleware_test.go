package webserver

import (
	"context"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/valyala/fasthttp"

	"github.com/InsideGallery/core/testutils"
)

func TestMiddleware(t *testing.T) {
	echoChain := func(next fiber.Handler) fiber.Handler {
		return func(c fiber.Ctx) error {
			_, err := c.Write([]byte("echoChain:"))
			testutils.Equal(t, err, nil)
			return next(c)
		}
	}
	handler := func(c fiber.Ctx) error {
		_, err := c.Write([]byte("handler"))
		return err
	}

	testcases := []struct {
		name    string
		chains  []Chain
		handler func(c fiber.Ctx) error
		result  string
	}{
		{
			name:    "should return expected string for echo sequence",
			chains:  []Chain{echoChain},
			handler: handler,
			result:  "echoChain:handler",
		},
		{
			name:    "should return expected string without any chains",
			chains:  []Chain{},
			handler: handler,
			result:  "handler",
		},
	}
	for _, test := range testcases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			m := NewMiddleware(test.chains...)
			fctx := &fasthttp.RequestCtx{}
			ctx := NewFiberApp("test").AcquireCtx(fctx)
			resp := ctx.Response()
			err := m.Then(test.handler)(ctx)
			testutils.Equal(t, err, nil)
			data := resp.Body()
			testutils.Equal(t, err, nil)
			testutils.Equal(t, string(data), test.result)
		})
	}
}

func TestMiddlewareMerge(t *testing.T) {
	m1 := NewMiddleware(func(h fiber.Handler) fiber.Handler {
		return h
	})
	m2 := NewMiddleware(func(h fiber.Handler) fiber.Handler {
		return h
	})
	m3 := NewMiddleware(func(h fiber.Handler) fiber.Handler {
		return h
	})
	m1.Merge(m2, m3)
	m2.chains[0] = nil
	for _, c := range m1.chains {
		testutils.NotEqual(t, c, nil)
	}
	testutils.Equal(t, len(m1.chains), 3)
	testutils.Equal(t, len(m2.chains), 1)
	testutils.Equal(t, len(m3.chains), 1)
}

func TestMiddlewareCopy(t *testing.T) {
	m1 := NewMiddleware(func(h fiber.Handler) fiber.Handler {
		return h
	})
	m2 := m1.Copy()
	m1.chains[0] = nil
	for _, c := range m2.chains {
		testutils.NotEqual(t, c, nil)
	}
	testutils.Equal(t, len(m2.chains), 1)
}

func TestRouteMiddleware(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		want string
	}{
		{
			name: "wraps route handler",
			want: "chain:handler",
		},
	}

	for _, test := range cases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			middleware := NewRouteMiddleware(func(next RouteHandler) RouteHandler {
				return func(ctx context.Context, req RouteRequest) (RouteResponse, error) {
					req.Body = append([]byte("chain:"), req.Body...)

					return next(ctx, req)
				}
			})

			handler := middleware.Then(func(_ context.Context, req RouteRequest) (RouteResponse, error) {
				return RouteResponse{Body: req.Body}, nil
			})

			response, err := handler(context.Background(), RouteRequest{Body: []byte("handler")})
			if err != nil {
				t.Fatalf("handler: %v", err)
			}

			if string(response.Body) != test.want {
				t.Fatalf("body = %q, want %q", response.Body, test.want)
			}
		})
	}
}
