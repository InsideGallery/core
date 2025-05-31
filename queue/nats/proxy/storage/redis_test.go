//go:build local_test
// +build local_test

package storage

import (
	"context"
	"sort"
	"testing"

	"github.com/InsideGallery/core/db/redis"
	"github.com/InsideGallery/core/testutils"
)

func TestRedis(t *testing.T) {
	svc, err := redis.Default()
	testutils.Equal(t, err, nil)
	testutils.Equal(t, svc != nil, true)

	r := NewRedis(context.TODO(), "testredsync", svc.Client)
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
