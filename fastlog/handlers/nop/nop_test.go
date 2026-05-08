package nop

import (
	"testing"

	"github.com/InsideGallery/core/testutils"
)

func TestWriterWriteCountsBytes(t *testing.T) {
	cases := []struct {
		name    string
		payload []byte
	}{
		{
			name:    "nil payload",
			payload: nil,
		},
		{
			name:    "text payload",
			payload: []byte("discarded log event"),
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			w := W{}

			written, err := w.Write(test.payload)
			if err != nil {
				t.Fatalf("write: %v", err)
			}

			testutils.Equal(t, written, len(test.payload))
		})
	}
}
