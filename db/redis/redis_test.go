//go:build local_test
// +build local_test

package redis

import (
	"context"
	"testing"
	"time"

	guuid "github.com/google/uuid"

	"github.com/InsideGallery/core/testutils"
)

func TestNewRedisClient(t *testing.T) {
	svc, err := Default()
	testutils.Equal(t, err, nil)
	testutils.Equal(t, svc != nil, true)

	ctx := context.Background()

	// set
	testutils.Equal(t, svc.Set(ctx, "foo", "bar", time.Second), nil)

	// check existed
	val, present, err := svc.Get(ctx, "foo")
	testutils.Equal(t, err, nil)
	testutils.Equal(t, present, true)
	testutils.Equal(t, val, "bar")

	// get not existed
	val, present, err = svc.Get(ctx, guuid.NewString())
	testutils.Equal(t, err, nil)
	testutils.Equal(t, present, false)
	testutils.Equal(t, val, "")

	time.Sleep(time.Second)

	// expired
	val, present, err = svc.Get(ctx, "foo")
	testutils.Equal(t, err, nil)
	testutils.Equal(t, present, false)
	testutils.Equal(t, val, "")
}
