package logstash

import (
	"log/slog"
	"testing"
)

func TestNewFromConfig(t *testing.T) {
	cases := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "invalid network returns error",
			cfg: Config{
				Host:    "localhost:4242",
				Network: "bad-network",
				Level:   slog.LevelInfo,
			},
			wantErr: true,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			writer, opts, err := NewFromConfig(test.cfg)
			if test.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}

				return
			}

			if err != nil {
				t.Fatalf("new from config: %v", err)
			}

			if writer == nil || opts == nil {
				t.Fatal("writer or options are nil")
			}
		})
	}
}
