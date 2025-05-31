package hll

import (
	"testing"

	"github.com/InsideGallery/core/testutils"
)

func TestRestore(t *testing.T) {
	h1, err := New()
	testutils.Equal(t, err, nil)

	for i := 0; i < 10000; i++ {
		err = h1.AddAny(i)
		testutils.Equal(t, err, nil)
	}

	testutils.Equal(t, h1.Count(), uint64(10000))

	dump := h1.ToBytes()

	h1, err = FromBytes(dump)
	testutils.Equal(t, err, nil)

	testutils.Equal(t, h1.Count(), uint64(10000))
}

func TestHLL(t *testing.T) {
	h1, err := New()
	testutils.Equal(t, err, nil)

	h2, err := New()
	testutils.Equal(t, err, nil)

	for i := 0; i < 10000; i++ {
		err = h1.AddAny(i)
		testutils.Equal(t, err, nil)
	}

	for i := 5000; i < 15000; i++ {
		err = h2.AddAny(i)
		testutils.Equal(t, err, nil)
	}

	c, err := h1.IntersectionCount(h2)
	testutils.Equal(t, err, nil)

	u, err := h1.UnionCount(h2)
	testutils.Equal(t, err, nil)

	testutils.Equal(t, c, uint64(5000))
	testutils.Equal(t, h1.Count(), uint64(10000))
	testutils.Equal(t, h2.Count(), uint64(10000))
	testutils.Equal(t, u, uint64(15000))
}
