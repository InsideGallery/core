# pki/diversify

Import path: `github.com/InsideGallery/core/pki/diversify`

`diversify` derives AES keys using the AN10922-style AES-CMAC diversification
flow used elsewhere in the PKI packages. It supports 128-bit, 192-bit, and
256-bit master keys.

## Main API

- `Key(masterKey, diversificationData)` returns a diversified key with the same
  length as the master key.
- `DiversifyKey(masterKey, diversificationData)` is the deprecated compatibility
  name for `Key`.
- `ErrWrongKeyLen` reports master keys that are not 16, 24, or 32 bytes long.
- `DiversityConstant128`, `DiversityConstant192_1`,
  `DiversityConstant192_2`, `DiversityConstant256_1`, and
  `DiversityConstant256_2` expose the constants used by the derivation.

The `diversificationData` argument should not include the diversity constant.
It should include the remaining caller-specific data, such as UID, application
ID, and system identifier.

## Usage

```go
package example

import "github.com/InsideGallery/core/pki/diversify"

func derive(masterKey, applicationID, systemID []byte) ([]byte, error) {
	data := make([]byte, 0, len(applicationID)+len(systemID))
	data = append(data, applicationID...)
	data = append(data, systemID...)

	return diversify.Key(masterKey, data)
}
```

## Security Notes

The function is deterministic for the same master key and diversification data.
Keep master keys secret and pass the diversification data in the exact byte
order expected by the external protocol you are interoperating with.
