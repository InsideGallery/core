package client

import (
	"testing"

	"github.com/InsideGallery/core/testutils"
)

func TestGetNATSConnectionConfigFromEnv(t *testing.T) {
	t.Setenv("CUSTOM_NATS_ADDR", "test")

	// success with custom variable
	c, err := GetNATSConnectionConfigFromEnv("custom_nats")
	testutils.Equal(t, err, nil)
	testutils.Equal(t, c.Addr, "test")

	// success with unknown/invalid prefix
	_, err = GetNATSConnectionConfigFromEnv("unknown_nats")
	testutils.Equal(t, err, nil)
}
