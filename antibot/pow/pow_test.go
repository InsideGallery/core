package pow

import (
	"testing"

	"github.com/InsideGallery/core/testutils"
)

func TestProofOfWork(t *testing.T) {
	pow := NewProofOfWork(4) // Difficulty: 4 leading zeros
	message := "Hello, world!"
	nonce, hash := pow.FindNonce(message)

	testutils.Equal(t, nonce, 4250)
	testutils.Equal(t, hash, "0000c3af42fc31103f1fdc0151fa747ff87349a4714df7cc52ea464e12dcd4e9")
}
