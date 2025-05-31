//go:build integration
// +build integration

package storage

import (
	"sort"
	"testing"

	"github.com/InsideGallery/core/server/instance"

	"github.com/InsideGallery/core/db/aerospike"

	"github.com/InsideGallery/core/testutils"
)

func TestAerospike(t *testing.T) {
	svc, err := aerospike.Default()
	testutils.Equal(t, err, nil)
	testutils.Equal(t, svc != nil, true)

	r := NewAerospike("transactions", instance.GetInstanceID(), svc)
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
