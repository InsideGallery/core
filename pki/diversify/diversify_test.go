package diversify

import (
	"encoding/hex"
	"testing"

	"github.com/InsideGallery/core/testutils"
)

var (
	DesFireAID                      = []byte{0x30, 0x42, 0xF5}
	SystemIdentifierForDiversifying = MustDecodeHEX("666F6F")
)

func MustDecodeHEX(str string) []byte {
	key, err := hex.DecodeString(str)
	if err != nil {
		panic(err)
	}

	return key
}

func TestDiversifyKey(t *testing.T) {
	key := make([]byte, 32)

	resKey, err := DiversifyKey(key, append(DesFireAID, SystemIdentifierForDiversifying...))
	testutils.Equal(t, err, nil)

	testutils.Equal(t,
		hex.EncodeToString(resKey),
		"b5054fa8b0b11852115732183532fdc87c2199e36fbcac70049de7c8cf5585d8",
	)

	key = make([]byte, 24)

	resKey, err = DiversifyKey(key, append(DesFireAID, SystemIdentifierForDiversifying...))
	testutils.Equal(t, err, nil)

	testutils.Equal(t,
		hex.EncodeToString(resKey),
		"43b1a4d765326ce0427b1374b27dbff6c3c24e94a1960409",
	)

	key = make([]byte, 16)

	resKey, err = DiversifyKey(key, append(DesFireAID, SystemIdentifierForDiversifying...))
	testutils.Equal(t, err, nil)

	testutils.Equal(t,
		hex.EncodeToString(resKey),
		"7ddc20a207ef0c7d7c7c40f36725035c",
	)
}
