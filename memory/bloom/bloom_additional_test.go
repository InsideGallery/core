package bloom

import (
	"errors"
	"testing"
)

func TestCountingFilterBoundaries(t *testing.T) {
	filter := &CountingFilter{
		filter:   newFilter(1, 1),
		counters: []byte{255},
	}

	filter.Add([]byte("overflow"))
	if filter.counters[0] != 255 {
		t.Fatalf("counter after overflow add = %d, want 255", filter.counters[0])
	}

	filter.Remove([]byte("underflow"))
	filter.Remove([]byte("underflow"))
	if filter.counters[0] != 253 {
		t.Fatalf("counter after removes = %d, want 253", filter.counters[0])
	}

	filter.Reset()
	if filter.counters[0] != 0 {
		t.Fatalf("counter after reset = %d, want 0", filter.counters[0])
	}

	filter.Remove([]byte("underflow"))
	if filter.counters[0] != 0 {
		t.Fatalf("counter after underflow remove = %d, want 0", filter.counters[0])
	}
}

func TestCountingFilterDecodeErrors(t *testing.T) {
	cases := []struct {
		name string
		data []byte
		want error
	}{
		{
			name: "empty dump",
			want: ErrEmptyDump,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got, err := NewCountingFromBytes(test.data)
			if err == nil {
				t.Fatal("expected error")
			}

			if got != nil {
				t.Fatalf("filter = %#v, want nil", got)
			}

			if test.want != nil && !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
}
