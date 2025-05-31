package utils

import (
	"testing"

	"github.com/InsideGallery/core/testutils"
)

func TestCRC32(t *testing.T) {
	testutils.Equal(t, CRC32("test1"), uint32(1409163093))
	testutils.Equal(t, CRC32("test2"), uint32(1085205665))
	testutils.Equal(t, CRC32("true"), uint32(151551613))
	testutils.Equal(t, CRC32("false"), uint32(118305666))
}
