# stdx/bytes

Import path: `github.com/InsideGallery/core/stdx/bytes`

## Overview

`stdx/bytes` provides byte, bit, little-endian, XOR, CRC, rotation, and padding helpers.

## Main APIs

- `GetByteLSB`, `GetBitLSB`, `LSBBytesToInt`, and `LSBBitValue` work with least-significant-byte ordering.
- `LeftNibble`, `RightNibble`, and `UnsignedByteToInt` extract byte parts and unsigned values.
- `XOR` and `XORAlt` return XORed byte slices.
- `JamCRC32` returns the little-endian IEEE CRC-32 with all bits inverted.
- `RotateLeft` and `RotateRight` rotate byte slices by the requested count.
- `PadMessageToBlocksize` pads non-aligned messages with `0x80` followed by zero bytes.

## Usage

```go
crc := bytesx.JamCRC32([]byte("123456789"))
rotated := bytesx.RotateLeft([]byte{1, 2, 3, 4}, 1)
value := bytesx.LSBBytesToInt([]byte{0x34, 0x12})

_ = crc
_ = rotated
_ = value
```

## Notes

The helpers do not read configuration. Functions that return slices allocate new result slices except
`PadMessageToBlocksize`, which returns the original message when its length is already a multiple of the block
size.
