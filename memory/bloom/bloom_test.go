package bloom

import (
	"testing"

	"github.com/InsideGallery/core/testutils"
)

var (
	foo = []byte("foo")
	bar = []byte("bar")
	baz = []byte("baz")
)

func TestCountingFilter(t *testing.T) {
	f := NewCounting(3000, 0.01)
	f.Add(foo)
	f.Add(foo)
	f.Remove(foo)

	if !f.Test(foo) {
		t.Error("foo not in bloom filter")
	}

	f.Remove(foo)

	if f.Test(foo) {
		t.Error("foo still in bloom filter")
	}
}

func TestDump(t *testing.T) {
	f := NewCounting(3000, 0.01)
	f.Add(foo)
	f.Add(bar)
	f.Remove(foo)

	testutils.Equal(t, f.Test(foo), false)
	testutils.Equal(t, f.Test(bar), true)

	dump, err := f.ToBytes()
	testutils.Equal(t, err, nil)

	f2, err := NewCountingFromBytes(dump)
	testutils.Equal(t, err, nil)

	testutils.Equal(t, f2.Test(foo), false)
	testutils.Equal(t, f2.Test(bar), true)
	testutils.Equal(t, f2.Test(baz), false)

	f2.Add(baz)

	testutils.Equal(t, f2.Test(baz), true)
}
