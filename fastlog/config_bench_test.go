package fastlog

import (
	"log/slog"
	"testing"
)

var benchmarkHandler slog.Handler

func BenchmarkConfigGetHandler(b *testing.B) {
	cases := []struct {
		name string
		cfg  Config
	}{
		{
			name: "nop_json",
			cfg: Config{
				Outputs: []string{"nop:json"},
				Level:   slog.LevelInfo,
			},
		},
		{
			name: "nop_json_with_middlewares",
			cfg: Config{
				Outputs:         []string{"nop:json"},
				Level:           slog.LevelInfo,
				Caller:          true,
				ErrorFormatting: true,
			},
		},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				handler, err := tc.cfg.GetHandler()
				if err != nil {
					b.Fatal(err)
				}

				benchmarkHandler = handler
			}
		})
	}
}
