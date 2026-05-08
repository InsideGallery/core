package nop

import "testing"

func TestNewFromConfig(t *testing.T) {
	cases := []struct {
		name    string
		payload []byte
	}{
		{
			name:    "empty payload",
			payload: nil,
		},
		{
			name:    "non-empty payload",
			payload: []byte("discarded"),
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			writer, opts, err := NewFromConfig()
			if err != nil {
				t.Fatalf("new writer: %v", err)
			}

			if opts == nil {
				t.Fatal("handler options are nil")
			}

			n, err := writer.Write(test.payload)
			if err != nil {
				t.Fatalf("write: %v", err)
			}

			if n != len(test.payload) {
				t.Fatalf("written = %d, want %d", n, len(test.payload))
			}
		})
	}
}

func TestNewCompatibility(t *testing.T) {
	cases := []struct {
		name string
	}{
		{name: "new wrapper returns writer"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			writer, opts, err := New()
			if err != nil {
				t.Fatalf("new writer: %v", err)
			}

			if writer == nil {
				t.Fatal("writer is nil")
			}

			if opts == nil {
				t.Fatal("handler options are nil")
			}
		})
	}
}
