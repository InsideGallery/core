//go:build unit
// +build unit

package stack

import (
	"testing"

	"github.com/InsideGallery/core/testutils"
)

func TestPrefix_PopLeft(t *testing.T) {
	var arr Stack[string]
	arr.Set([]string{
		"test123",
		"test124",
		"test125",
	})

	left := arr.PopLeft()
	testutils.Equal(t, left, "test123")

	testutils.Equal(t, arr.Len(), 2)
}
