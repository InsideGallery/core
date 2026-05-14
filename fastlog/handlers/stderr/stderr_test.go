package stderr

import (
	"log/slog"
	"os"
	"testing"
)

func TestNewReturnsStderrWithConfiguredLevel(t *testing.T) {
	t.Setenv("STDERR_LEVEL", "DEBUG")

	writer, opts, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	if writer != os.Stderr {
		t.Fatalf("writer = %v, want os.Stderr", writer)
	}

	if opts == nil {
		t.Fatal("expected handler options")
	}

	if opts.Level.Level() != slog.LevelDebug {
		t.Fatalf("Level = %v, want DEBUG", opts.Level.Level())
	}
}

func TestNewFallsBackToStderrWhenEnvIsInvalid(t *testing.T) {
	t.Setenv("STDERR_LEVEL", "not-a-level")

	writer, opts, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	if writer != os.Stderr {
		t.Fatalf("writer = %v, want os.Stderr", writer)
	}

	if opts != nil {
		t.Fatalf("opts = %+v, want nil fallback options", opts)
	}
}
