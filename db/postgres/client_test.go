//go:build integration
// +build integration

package postgres

import (
	"testing"

	"github.com/InsideGallery/core/testutils"
)

func TestDefault(t *testing.T) {
	// t.Skip() // TODO Fix and uncomment.
	c, err := Get()
	testutils.Equal(t, err, ErrConnectionIsNotSet)
	testutils.Equal(t, c == nil, true)

	c, err = Default()
	testutils.Equal(t, err, nil)
	testutils.Equal(t, c != nil, true)

	row := c.QueryRowx("select 1")

	var v interface{}
	testutils.Equal(t, row.Scan(&v), nil)
	testutils.Equal(t, v, int64(1))
}
