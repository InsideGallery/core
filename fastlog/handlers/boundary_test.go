package handlers

import (
	"context"
	"log/slog"
	"testing"
)

func TestResolve(t *testing.T) {
	RegisterHandler("boundary-test", slog.NewTextHandler(discardWriter{}, nil))

	cases := []struct {
		name    string
		options Options
		wantErr bool
	}{
		{
			name: "registered handler",
			options: Options{
				Kind:   "boundary-test",
				Format: FormatText,
			},
		},
		{
			name: "missing handler",
			options: Options{
				Kind:   "missing-boundary-test",
				Format: FormatText,
			},
			wantErr: true,
		},
	}

	for _, test := range cases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			got, err := Resolve(test.options)
			if test.wantErr {
				if err == nil {
					t.Fatal("Resolve() expected error")
				}

				return
			}

			if err != nil {
				t.Fatalf("Resolve() error: %v", err)
			}

			if got.Handler == nil {
				t.Fatal("Handler is nil")
			}
		})
	}
}

type discardWriter struct{}

func (discardWriter) Write(data []byte) (int, error) {
	return len(data), nil
}

func (discardWriter) Enabled(context.Context, slog.Level) bool {
	return true
}
