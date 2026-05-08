package datadog

import (
	"context"
	"log/slog"
	"testing"
	"time"
)

func TestNewHandler(t *testing.T) {
	cases := []struct {
		name string
		cfg  Config
	}{
		{
			name: "explicit config creates handler",
			cfg: Config{
				Service:  "unit",
				Endpoint: "datadoghq.eu",
				Timeout:  time.Millisecond,
				Level:    slog.LevelInfo,
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			handler, err := NewHandler(context.Background(), test.cfg)
			if err != nil {
				t.Fatalf("new handler: %v", err)
			}

			if handler == nil {
				t.Fatal("handler is nil")
			}
		})
	}
}
