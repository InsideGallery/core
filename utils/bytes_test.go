package utils

import (
	"testing"

	"github.com/InsideGallery/core/testutils"
)

func TestXOR(t *testing.T) {
	x := XOR([]byte{1, 2, 3}, []byte{2, 4, 5})
	x2 := XORAlt([]byte{1, 2, 3}, []byte{2, 4, 5})

	testutils.Equal(t, x, x2)
}
