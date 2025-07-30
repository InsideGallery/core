package lru

import (
	"strconv"
	"testing"

	"github.com/InsideGallery/core/testutils"
)

func TestLRU(t *testing.T) {
	l := NewLRUCache[any](10)

	for i := 0; i < 20; i++ {
		k := strconv.Itoa(i)
		l.Put("key"+k, "val"+k)
	}

	v, ok := l.Get("test")
	testutils.Equal(t, ok, false)
	testutils.Equal(t, v, nil)

	v, ok = l.Get("key12")
	testutils.Equal(t, ok, true)
	testutils.Equal(t, v, "val12")

	for i := 20; i < 28; i++ {
		k := strconv.Itoa(i)
		l.Put("key"+k, "val"+k)
	}

	v, ok = l.Get("key12")
	testutils.Equal(t, ok, true)
	testutils.Equal(t, v, "val12")

	v, ok = l.Get("key22")
	testutils.Equal(t, ok, true)
	testutils.Equal(t, v, "val22")

	v, ok = l.Get("key13")
	testutils.Equal(t, ok, false)
	testutils.Equal(t, v, nil)
}
