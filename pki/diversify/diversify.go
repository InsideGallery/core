package diversify

import (
	"errors"

	"github.com/InsideGallery/core/pki/aescmac"
	"github.com/InsideGallery/core/stdx/bytes"
)

var (
	DiversityConstant128   = []byte{0x01}
	DiversityConstant192_1 = []byte{0x11}
	DiversityConstant192_2 = []byte{0x12}
	DiversityConstant256_1 = []byte{0x41}
	DiversityConstant256_2 = []byte{0x42}
)

var ErrWrongKeyLen = errors.New("key must be 16, 24, or 32 bytes long")

const (
	aesKeySize128    = 16
	aesKeySize192    = 24
	aesKeySize256    = 32
	cmacHalfBlockLen = 8
)

// Key diversifies keys according to the AES standards in AN10922 for 128, 196, and 256 bit keys.
// A wrong-sized key will throw and IllegalArgumentException.
// The diversificationData should *not* include the diversity constant,
// but should include everything else (uid, application id, and system identifier).
func Key(masterKey, diversificationData []byte) ([]byte, error) {
	// NOTE - we are not including the padblock because the CMAC function already does it
	switch len(masterKey) {
	case aesKeySize128:
		keyData, err := diversifiedSum(masterKey, DiversityConstant128, diversificationData)
		if err != nil {
			return nil, err
		}

		return keyData, nil
	case aesKeySize192:
		a, err := diversifiedSum(masterKey, DiversityConstant192_1, diversificationData)
		if err != nil {
			return nil, err
		}

		b, err := diversifiedSum(masterKey, DiversityConstant192_2, diversificationData)
		if err != nil {
			return nil, err
		}

		keyData := make([]byte, 0, aesKeySize192)
		keyData = append(keyData, a[:cmacHalfBlockLen]...)
		keyData = append(keyData, bytes.XOR(a[cmacHalfBlockLen:], b[:cmacHalfBlockLen])...)
		keyData = append(keyData, b[cmacHalfBlockLen:]...)

		return keyData, nil
	case aesKeySize256:
		a, err := diversifiedSum(masterKey, DiversityConstant256_1, diversificationData)
		if err != nil {
			return nil, err
		}

		b, err := diversifiedSum(masterKey, DiversityConstant256_2, diversificationData)
		if err != nil {
			return nil, err
		}

		return append(a, b...), nil
	default:
		return nil, ErrWrongKeyLen
	}
}

// DiversifyKey diversifies keys according to the AES standards in AN10922 for 128, 196, and 256 bit keys.
//
// Deprecated: use Key.
func DiversifyKey(masterKey, diversificationData []byte) ([]byte, error) { //nolint:revive
	return Key(masterKey, diversificationData)
}

func diversifiedSum(masterKey, constant, diversificationData []byte) ([]byte, error) {
	data := make([]byte, 0, len(constant)+len(diversificationData))
	data = append(data, constant...)
	data = append(data, diversificationData...)

	return aescmac.Sum(masterKey, data)
}
