package middlewares

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

const HeaderXTraceID = "X-Trace-ID"

var (
	operation = "HTTP"
	otelOpts  = []otelhttp.Option{
		otelhttp.WithSpanNameFormatter(DefaultSpanNameFormatter),
		otelhttp.WithFilter(DefaultFilter),
	}

	reID       = regexp.MustCompile(`^\d+$`)
	reResource = regexp.MustCompile(`^[a-zA-Z0-9\-]+\.\w{2,4}$`) // .css, .js, .png, .jpeg, etc.
	reUUID     = regexp.MustCompile(`^[a-f\d]{4}(?:[a-f\d]{4}-){4}[a-f\d]{12}$`)

	decreasePathCardinality = func(path string) string {
		var b strings.Builder

		path = strings.TrimLeft(path, "/")
		pathParts := strings.Split(path, "/")
		for _, part := range pathParts {
			b.WriteString("/")

			p := part
			if reID.MatchString(part) {
				p = ":id:"
			} else if reResource.MatchString(part) {
				p = ":resource:"
			} else if reUUID.MatchString(part) {
				p = ":uuid:"
			}
			b.WriteString(p)
		}

		return b.String()
	}

	DefaultSpanNameFormatter = func(_ string, r *http.Request) string {
		var b strings.Builder

		b.WriteString(r.Method)
		b.WriteString(":")
		b.WriteString(decreasePathCardinality(r.URL.Path))

		return b.String()
	}

	DefaultFilter = func(r *http.Request) bool {
		if k, ok := r.Header["Upgrade"]; ok {
			for _, v := range k {
				if v == "websocket" {
					return false
				}
			}
		}

		return !(r.Method == http.MethodGet && strings.HasPrefix(r.URL.RequestURI(), "/health"))
	}
)

func Telemetry() func(next fiber.Handler) fiber.Handler {
	instance := func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(next, operation, otelOpts...)
	}

	w := httptest.NewRecorder() // not use so can do like this

	return func(next fiber.Handler) fiber.Handler {
		return func(c *fiber.Ctx) error {
			req, err := http.NewRequest(http.MethodGet, "", nil)
			if err == nil {
				_ = fasthttpadaptor.ConvertRequest(c.Context(), req, true)
			}

			instance(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if span := trace.SpanFromContext(r.Context()); span != nil {
					traceID := span.SpanContext().TraceID().String()
					c.Response().Header.Set(HeaderXTraceID, traceID)
				}

				err = next(c)
				if err != nil {
					slog.Default().Error("Error call next function", "err", err)
				}

				w.WriteHeader(c.Context().Response.StatusCode())

				_, err = w.Write(c.Context().Response.Body())
				if err != nil {
					slog.Default().Error("Error write response", "err", err)
				}
			})).ServeHTTP(w, req)

			return nil
		}
	}
}
