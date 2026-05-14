//go:build !fastlog_minimal

package all_test

import (
	"log/slog"
	"testing"

	_ "github.com/InsideGallery/core/fastlog/all"

	"github.com/InsideGallery/core/fastlog/handlers"
)

func TestAllRegistersFastlogHandlers(t *testing.T) {
	t.Setenv("DATADOG_API_KEY", "unit-test")

	for _, kind := range []string{"datadog", "nop", "otel", "stderr"} {
		t.Run(kind, func(t *testing.T) {
			if _, err := handlers.Get(kind, handlers.FormatJSON, slog.LevelInfo); err != nil {
				t.Fatalf("handlers.Get(%q) error: %v", kind, err)
			}
		})
	}
}
