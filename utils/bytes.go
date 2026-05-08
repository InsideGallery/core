// Package utils is a legacy aggregate of byte, string, hash, context, password,
// slice, semver, and tokenizer helpers.
//
// New helpers should live in the package that owns the concept. Existing exports
// remain for compatibility.
package utils

import (
	"hash/crc32"

	"github.com/go-faster/xor"
)

const (
	bitsPerByte       = 8
	nibbleBitSize     = 4
	leftNibbleMask    = 0xF0
	rightNibbleMask   = 0x0F
	unsignedByteMask  = 0xff
	crcByteIndexTwo   = 2
	crcByteIndexThree = 3
	messagePadStart   = 0x80
)

func GetByteLSB(value int64, byteNumber int) byte {
	for byteNumber > 0 {
		value >>= bitsPerByte
		byteNumber--
	}

	return byte(value)
}

func GetBitLSB(b byte, bit int) bool {
	return b&(1<<bit) != 0
}

func LeftNibble(input byte) int {
	return (int(input) & leftNibbleMask) >> nibbleBitSize
}

func RightNibble(input byte) int {
	return int(input) & rightNibbleMask
}

func UnsignedByteToInt(b byte) int {
	return int(b) & unsignedByteMask
}

func XOR(a1, a2 []byte) []byte {
	dst := make([]byte, len(a1))
	xor.Bytes(dst, a1, a2)

	return dst
}

func XORAlt(a1, a2 []byte) []byte {
	dst := make([]byte, len(a1))

	for i := range dst {
		dst[i] = a1[i] ^ a2[i]
	}

	return dst
}

func LSBBytesToInt(data []byte) int {
	multiplier := 1
	value := 0

	for i := 0; i < len(data); i++ {
		value += UnsignedByteToInt(data[i]) * multiplier
		multiplier *= 256
	}

	return value
}

func LSBBitValue(bitIdx int, isSet bool) byte {
	if !isSet {
		return 0
	}

	return 1 << bitIdx
}

func JamCRC32(value []byte) []byte {
	result := crc32.ChecksumIEEE(value)
	basicCRC := []byte{
		GetByteLSB(int64(result), 0),
		GetByteLSB(int64(result), 1),
		GetByteLSB(int64(result), crcByteIndexTwo),
		GetByteLSB(int64(result), crcByteIndexThree),
	}

	jamCRC := XOR(basicCRC,
		[]byte{
			0xff,
			0xff,
			0xff,
			0xff,
		},
	)

	return jamCRC
}

// RotateLeft byte-rotates a byte array left by the given number of rotations
func RotateLeft(val []byte, rotations int) []byte {
	result := make([]byte, len(val))

	for i := 0; i < len(val); i++ {
		var newIdx int
		if i < rotations {
			newIdx = len(val) - (rotations + i)
		} else {
			newIdx = i - rotations
		}

		result[newIdx] = val[i]
	}

	return result
}

// RotateRight byte-rotates a byte array left by the given number of rotations
func RotateRight(val []byte, rotations int) []byte {
	result := make([]byte, len(val))

	for i := 0; i < len(val); i++ {
		var newIdx int
		if i >= (len(val) - rotations) {
			newIdx = i - len(val) + rotations
		} else {
			newIdx = i + rotations
		}

		result[newIdx] = val[i]
	}

	return result
}

func PadMessageToBlocksize(message []byte, blocksize int) []byte {
	reminder := len(message) % blocksize
	compleBlocks := len(message) / blocksize

	if reminder == 0 {
		return message
	}

	result := make([]byte, (compleBlocks+1)*blocksize)
	copy(result, message)
	result[len(message)] = messagePadStart

	return result
}
