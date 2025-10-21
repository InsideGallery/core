package aescmac

import (
	"encoding/hex"
	"testing"

	"github.com/InsideGallery/core/testutils"
)

func TestAESCMAC(t *testing.T) {
	key := make([]byte, 16)

	hash, err := NewCMAC(key)
	testutils.Equal(t, err, nil)

	res := hash.Sum([]byte("test string"))
	testutils.Equal(t, hex.EncodeToString(res), "7465737420737472696e674387c14b46ef7e176dceefa862d72ff9")
}
