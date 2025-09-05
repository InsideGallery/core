package mathutils

import (
	"testing"

	"github.com/InsideGallery/core/testutils"
)

func TestCantor(t *testing.T) {
	var (
		ev1 uint64 = 2
		ev2 uint64 = 1
	)

	v := CantorPair(ev1, ev2)
	testutils.Equal(t, v, uint64(7))

	v1, v2 := CantorUnpair(v)
	testutils.Equal(t, v1, ev1)
	testutils.Equal(t, v2, ev2)
}
