package utils

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"math/big"
	mrand "math/rand/v2"
	"strings"
	"time"
	"unsafe"

	"github.com/sirbu/golang-common/hash/crc16"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ByteSliceToString cast given bytes to string, without allocation memory
func ByteSliceToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b)) //nolint
}

// GetUniqueID return unique id
func GetUniqueID() string {
	return primitive.NewObjectID().Hex()
}

// GetShortID return short id
func GetShortID() ([]byte, error) {
	b := make([]byte, 2) //nolint:mnd

	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	r := make([]byte, 4) //nolint:mnd
	binary.BigEndian.PutUint32(r, uint32(time.Now().Nanosecond()))
	src := append(b, r...) //nolint:gocritic
	dst := make([]byte, hex.EncodedLen(len(src)))
	hex.Encode(dst, src)

	return dst, nil
}

// GetTinyID return tiny id
func GetTinyID() ([]byte, error) {
	b := make([]byte, 4) //nolint:mnd

	_, err := rand.Read(b) //nolint:gosec
	if err != nil {
		return nil, err
	}

	r := make([]byte, 4) //nolint:mnd
	// time.Now().UnixNano()
	binary.BigEndian.PutUint32(r, uint32(time.Now().Nanosecond()))
	b = append(b, r...)
	val := binary.BigEndian.Uint64(b)

	return []byte(big.NewInt(int64(val)).Text(62))[5:], nil //nolint:mnd
}

// Between function to get content between two keys
func Between(data string, keys ...string) string {
	var key1, key2 string

	switch {
	case len(keys) == 1: //nolint:mnd
		key1 = keys[0]
		key2 = keys[0]
	case len(keys) >= 2: //nolint:mnd
		key1 = keys[0]
		key2 = keys[1]
	default:
		return ""
	}

	if key1 == "" || key2 == "" {
		return ""
	}

	s := strings.Index(data, key1)
	if s <= -1 {
		return ""
	}

	s += len(key1)

	e := strings.Index(data[s:], key2)
	if e <= -1 {
		return ""
	}

	return strings.TrimSpace(data[s : s+e])
}

// SafeGet return value of pointer, and return default value if it nil
func SafeGet[T any](ptr *T, defaultValue T) T {
	if ptr == nil {
		return defaultValue
	}

	return *ptr
}

func MaskField(str string, keepUnmaskedFront int, keepUnmaskedEnd int) string {
	var result strings.Builder

	size := len(str)
	defaultResult := strings.Repeat("*", size)

	if size <= (keepUnmaskedFront+keepUnmaskedEnd)*2 {
		return defaultResult
	}

	_, err := result.WriteString(str[:keepUnmaskedFront])
	if err != nil {
		return defaultResult
	}

	_, err = result.WriteString(strings.Repeat("*", size-keepUnmaskedFront-keepUnmaskedEnd))
	if err != nil {
		return defaultResult
	}

	_, err = result.WriteString(str[size-keepUnmaskedEnd:])
	if err != nil {
		return defaultResult
	}

	return result.String()
}

func SplitByChunks(s string, chunkSize int) []string {
	if chunkSize <= 0 {
		return nil
	}

	var chunks []string

	for i := 0; i < len(s); i += chunkSize {
		end := i + chunkSize
		if end > len(s) {
			end = len(s)
		}

		chunks = append(chunks, s[i:end])
	}

	return chunks
}

const letterBytes = "abcdefghijklmnopqrstuvwxyz"

func RandStringBytes(n int) string {
	if n <= 0 {
		return ""
	}

	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[mrand.IntN(len(letterBytes))] // nolint:gosec
	}

	return string(b)
}

func HashName(name string) string {
	name = strings.ToLower(name)
	val := crc16.Checksum(crc16.XModem, []byte(name))

	r := make([]byte, 2) //nolint:mnd
	binary.BigEndian.PutUint16(r, val)

	return name[:1] + hex.EncodeToString(r)
}
