package diversify

import (
	"errors"

	"github.com/InsideGallery/core/pki/aescmac"
	"github.com/InsideGallery/core/utils"
)

var (
	DiversityConstant128   = []byte{0x01}
	DiversityConstant192_1 = []byte{0x11}
	DiversityConstant192_2 = []byte{0x12}
	DiversityConstant256_1 = []byte{0x41}
	DiversityConstant256_2 = []byte{0x42}
)

var ErrWrongKeyLen = errors.New("key must be 16, 24, or 32 bytes long")

// DiversifyKey diversifies keys according to the AES standards in AN10922 for 128, 196, and 256 bit keys.
// A wrong-sized key will throw and IllegalArgumentException.
// The diversificationData should *not* include the diversity constant,
// but should include everything else (uid, application id, and system identifier).
func DiversifyKey(masterKey, diversificationData []byte) ([]byte, error) {
	// NOTE - we are not including the padblock because the CMAC function already does it
	switch len(masterKey) {
	case 16:
		keyData, err := aescmac.Sum(masterKey, append(DiversityConstant128, diversificationData...))
		if err != nil {
			return nil, err
		}

		return keyData, nil
	case 24:
		a, err := aescmac.Sum(masterKey, append(DiversityConstant192_1, diversificationData...))
		if err != nil {
			return nil, err
		}

		b, err := aescmac.Sum(masterKey, append(DiversityConstant192_2, diversificationData...))
		if err != nil {
			return nil, err
		}

		keyData := append(a[0:8], append(utils.XOR(a[8:16], b[0:8]), b[8:16]...)...)

		return keyData, nil
	case 32:
		a, err := aescmac.Sum(masterKey, append(DiversityConstant256_1, diversificationData...))
		if err != nil {
			return nil, err
		}

		b, err := aescmac.Sum(masterKey, append(DiversityConstant256_2, diversificationData...))
		if err != nil {
			return nil, err
		}

		return append(a, b...), nil
	default:
		return nil, ErrWrongKeyLen
	}
}
