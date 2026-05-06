package middlewares

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
)

const (
	httpRequestDuration = "http.request.duration"
	httpRequestCount    = "http.request.count"
	httpRequestError    = "http.request.error"
	requestCountValue   = 1
	unmatchedRoute      = "unmatched"
)

type metricRecorder interface {
	Count(name string, value int64, tags []string) error
	Distribution(name string, value float64, tags []string) error
}

// Metrics returns a Fiber middleware that records HTTP request metrics.
func Metrics(client metricRecorder) fiber.Handler {
	return func(ctx fiber.Ctx) error {
		start := time.Now()

		err := ctx.Next()

		duration := float64(time.Since(start).Milliseconds())
		status := responseStatus(ctx, err)
		tags := requestMetricTags(ctx, status, err)

		recordMetricDistribution(client, httpRequestDuration, duration, tags)
		recordMetricCount(client, httpRequestCount, requestCountValue, tags)

		if status >= http.StatusInternalServerError {
			recordMetricCount(client, httpRequestError, requestCountValue, tags)
		}

		return err
	}
}

func responseStatus(ctx fiber.Ctx, err error) int {
	if err == nil {
		return ctx.Response().StatusCode()
	}

	var fiberError *fiber.Error
	if errors.As(err, &fiberError) {
		return fiberError.Code
	}

	return http.StatusInternalServerError
}

func requestMetricTags(ctx fiber.Ctx, status int, err error) []string {
	return []string{
		"method:" + ctx.Method(),
		"route:" + requestRoute(ctx, err),
		"status_code:" + strconv.Itoa(status),
	}
}

func requestRoute(ctx fiber.Ctx, err error) string {
	if errors.Is(err, fiber.ErrNotFound) || errors.Is(err, fiber.ErrMethodNotAllowed) {
		return unmatchedRoute
	}

	route := ctx.Route().Path
	if route == "" {
		return unmatchedRoute
	}

	return route
}

func recordMetricCount(client metricRecorder, name string, value int64, tags []string) {
	if client == nil {
		return
	}

	if err := client.Count(name, value, tags); err != nil {
		slog.Warn("record http metric failed", "metric", name, "error", err)
	}
}

func recordMetricDistribution(client metricRecorder, name string, value float64, tags []string) {
	if client == nil {
		return
	}

	if err := client.Distribution(name, value, tags); err != nil {
		slog.Warn("record http metric failed", "metric", name, "error", err)
	}
}
