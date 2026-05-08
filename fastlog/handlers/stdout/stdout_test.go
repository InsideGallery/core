package stdout

import (
	"log/slog"
	"os"
	"testing"

	"github.com/InsideGallery/core/testutils"
)

func TestNewFromConfigWriterIdentity(t *testing.T) {
	cases := []struct {
		name      string
		cfg       Config
		wantLevel slog.Level
	}{
		{
			name: "returns stdout with configured level",
			cfg: Config{
				Level: slog.LevelDebug,
			},
			wantLevel: slog.LevelDebug,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			writer, opts, err := NewFromConfig(test.cfg)
			if err != nil {
				t.Fatalf("new from config: %v", err)
			}

			if writer != os.Stdout {
				t.Fatal("writer is not stdout")
			}

			if opts == nil || opts.Level == nil {
				t.Fatal("handler options are incomplete")
			}

			testutils.Equal(t, opts.Level.Level(), test.wantLevel)
		})
	}
}
