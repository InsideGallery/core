package otel

import (
	"log/slog"
	"testing"

	"github.com/agoda-com/opentelemetry-go/otelslog"
	sdk "github.com/agoda-com/opentelemetry-logs-go/sdk/logs"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestNewHandler(t *testing.T) {
	cases := []struct {
		name string
		opts *otelslog.HandlerOptions
	}{
		{
			name: "explicit provider creates handler",
			opts: &otelslog.HandlerOptions{
				Level: slog.LevelInfo,
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			provider := &LoggerProvider{
				LoggerProvider: sdk.NewLoggerProvider(),
				TracerProvider: sdktrace.NewTracerProvider(),
			}

			handler := NewHandler(provider, test.opts)
			if handler == nil {
				t.Fatal("handler is nil")
			}
		})
	}
}
