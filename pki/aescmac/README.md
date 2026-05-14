# pki/aescmac

Import path: `github.com/InsideGallery/core/pki/aescmac`

`aescmac` implements AES-CMAC and small byte helpers used by key
diversification code. It returns standard `hash.Hash` values for incremental
MAC calculation and also provides a one-shot `Sum` helper.

## Main API

- `NewCMAC(key)` returns a `hash.Hash` for AES-CMAC. Keys must be 16, 24, or 32
  bytes.
- `Sum(key, data)` computes a one-shot AES-CMAC value.
- `ErrUnsupportedKeySize` reports unsupported key sizes.
- `Xor(a, b)` returns the byte-wise XOR of equal-length slices and returns nil
  on length mismatch.
- `ShiftLeft(data)` shifts a byte slice left by one bit with carry propagation.
- `Padding(data)` appends the CMAC padding marker and zero padding as needed.
- `ErrAlreadyFinished` and `ErrXorLengthMismatch` are exported for compatibility;
  the current `hash.Hash` usage does not normally surface them from `Sum`.

## Usage

```go
package example

import "github.com/InsideGallery/core/pki/aescmac"

func mac(data []byte) ([]byte, error) {
	key := make([]byte, 16)

	return aescmac.Sum(key, data)
}
```

For incremental input:

```go
h, err := aescmac.NewCMAC(key)
if err != nil {
	return nil, err
}
if _, err := h.Write(part1); err != nil {
	return nil, err
}
if _, err := h.Write(part2); err != nil {
	return nil, err
}
tag := h.Sum(nil)
```

## Security Notes

The implementation is validated by tests against RFC 4493 vectors. Compare MAC
tags with a constant-time comparison, and keep CMAC keys in secret storage.
