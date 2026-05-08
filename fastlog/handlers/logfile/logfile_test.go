package logfile

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/InsideGallery/core/testutils"
)

func TestNewFromConfigCreatesAppendableFile(t *testing.T) {
	cases := []struct {
		name      string
		existing  []byte
		payload   []byte
		wantLevel slog.Level
		wantFile  string
	}{
		{
			name:      "appends payload to existing file",
			existing:  []byte("existing"),
			payload:   []byte("-new"),
			wantLevel: slog.LevelWarn,
			wantFile:  "existing-new",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			fileName := filepath.Join(t.TempDir(), "log.txt")
			if err := os.WriteFile(fileName, test.existing, 0o600); err != nil {
				t.Fatalf("seed log file: %v", err)
			}

			writer, opts, err := NewFromConfig(Config{
				Name:  fileName,
				Level: test.wantLevel,
			})
			if err != nil {
				t.Fatalf("new from config: %v", err)
			}

			if opts == nil || opts.Level == nil {
				t.Fatal("handler options are incomplete")
			}

			written, err := writer.Write(test.payload)
			if err != nil {
				t.Fatalf("write log file: %v", err)
			}

			testutils.Equal(t, written, len(test.payload))
			testutils.Equal(t, opts.Level.Level(), test.wantLevel)

			closer, ok := writer.(interface{ Close() error })
			if !ok {
				t.Fatal("writer is not closable")
			}

			if err := closer.Close(); err != nil {
				t.Fatalf("close log file: %v", err)
			}

			data, err := os.ReadFile(fileName)
			if err != nil {
				t.Fatalf("read log file: %v", err)
			}

			testutils.Equal(t, string(data), test.wantFile)
		})
	}
}
