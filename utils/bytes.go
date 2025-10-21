package utils

import (
	"hash/crc32"

	"github.com/go-faster/xor"
)

func GetByteLSB(value int64, byteNumber int) byte {
	for byteNumber > 0 {
		value = value >> 8
		byteNumber--
	}

	return byte(value)
}

func GetBitLSB(b byte, bit int) bool {
	return b&(1<<bit) != 0
}

func LeftNibble(input byte) int {
	return (int(input) & 0xF0) >> 4
}

func RightNibble(input byte) int {
	return int(input) & 0x0F
}

func UnsignedByteToInt(b byte) int {
	return int(b) & 0xff
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
	// var res []byte
	// basicCRC := binary.LittleEndian.AppendUint32(res, crc32.ChecksumIEEE(value))
	result := crc32.ChecksumIEEE(value)
	basicCRC := []byte{
		GetByteLSB(int64(result), 0),
		GetByteLSB(int64(result), 1),
		GetByteLSB(int64(result), 2),
		GetByteLSB(int64(result), 3),
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
		newIdx := i
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
	result[len(message)] = byte(0x80)

	return result
}
