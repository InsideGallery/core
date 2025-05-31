//go:build unit
// +build unit

package storage

import (
	"sort"
	"testing"

	"github.com/InsideGallery/core/testutils"
)

func TestMemory(t *testing.T) {
	r := NewMemory()
	testutils.Equal(t, r.Add("test", "abc"), nil)
	testutils.Equal(t, r.Add("test2", "abc2"), nil)
	testutils.Equal(t, r.Size("test"), 1)
	testutils.Equal(t, r.Size("test2"), 1)
	ids := r.GetIDs()
	sort.Strings(ids)
	testutils.Equal(t, ids, []string{"abc", "abc2"})
	testutils.Equal(t, r.Delete("test", "abc"), nil)
	testutils.Equal(t, r.Size("test"), 0)
	testutils.Equal(t, r.GetKeys("test2"), []string{"abc2"})
	testutils.Equal(t, r.GetIDs(), []string{"abc2"})
	testutils.Equal(t, r.Add("test", "abc3"), nil)
	testutils.Equal(t, r.Add("test2", "abc3"), nil)
	ids = r.GetIDs()
	sort.Strings(ids)
	testutils.Equal(t, ids, []string{"abc2", "abc3"})
	testutils.Equal(t, r.DeleteByID("abc3"), nil)
	testutils.Equal(t, r.GetIDs(), []string{"abc2"})
}
