package dataconv

import (
	"bytes"
	"testing"

	"github.com/InsideGallery/core/testutils"
)

var s []byte

func TestEncodeString(t *testing.T) {
	str := "Hello world!"
	b := NewBinaryEncoder()
	testutils.Equal(t, b.Encode(str), nil)
	data := b.Bytes()
	testutils.Equal(t, data, []byte{0, 0, 0, 12, 72, 101, 108, 108, 111, 32, 119, 111, 114, 108, 100, 33})

	d := NewBinaryDecoder(data)
	var result string
	err := d.Decode(&result)
	testutils.Equal(t, err, nil)
	testutils.Equal(t, result, str)
}

func TestEncodeUint(t *testing.T) {
	b := NewBinaryEncoder()
	testutils.Equal(t, b.Encode(uint8(1)), nil)
	testutils.Equal(t, b.Encode(uint16(1245)), nil)
	testutils.Equal(t, b.Encode(uint32(1245678)), nil)
	testutils.Equal(t, b.Encode(uint64(124567891011)), nil)
	testutils.Equal(t, b.Encode([]byte("test")), nil)
	testutils.Equal(t, b.Encode(12.56), nil)
	testutils.Equal(t, b.Encode(float32(3.14)), nil)
	data := b.Bytes()

	d := NewBinaryDecoder(data)
	var tuint8 uint8
	err := d.Decode(&tuint8)
	testutils.Equal(t, err, nil)
	testutils.Equal(t, tuint8, uint8(1))
	var tuint16 uint16
	err = d.Decode(&tuint16)
	testutils.Equal(t, err, nil)
	testutils.Equal(t, tuint16, uint16(1245))
	var tuint32 uint32
	err = d.Decode(&tuint32)
	testutils.Equal(t, err, nil)
	testutils.Equal(t, tuint32, uint32(1245678))
	var tuint64 uint64
	err = d.Decode(&tuint64)
	testutils.Equal(t, err, nil)
	testutils.Equal(t, tuint64, uint64(124567891011))
	var tbytes []byte
	err = d.Decode(&tbytes)
	testutils.Equal(t, err, nil)
	testutils.Equal(t, tbytes, []byte("test"))
	var tfloat64 float64
	err = d.Decode(&tfloat64)
	testutils.Equal(t, err, nil)
	testutils.Equal(t, tfloat64, float64(12.56))
	var tfloat32 float32
	err = d.Decode(&tfloat32)
	testutils.Equal(t, err, nil)
	testutils.Equal(t, tfloat32, float32(3.14))
}

/*
BenchmarkBinaryAppend-16        194472022                6.226 ns/op
BenchmarkBinaryCopy-16          1000000000               0.4647 ns/op
BenchmarkBinaryDecoder-16       32935569                37.94 ns/op
BenchmarkDecode-16              59279108                22.54 ns/op
BenchmarkBinaryEncoder-16       17599882                69.00 ns/op
BenchmarkEncode-16              18019960                71.73 ns/op
*/

var (
	localStr   string
	localBytes []byte
)

func BenchmarkBinaryAppend(b *testing.B) {
	t := []byte{124, 123, 212, 34, 12, 21, 1, 45, 76}
	l := make([]byte, 0, len(t))
	for i := 0; i < b.N; i++ {
		l = append(l, t...)
	}
	s = l
}

func BenchmarkBinaryCopy(b *testing.B) {
	t := []byte{124, 123, 212, 34, 12, 21, 1, 45, 76}
	l := make([]byte, len(t))
	for i := 0; i < b.N; i++ {
		copy(l, t)
	}
	s = l
}

func BenchmarkBinaryDecoder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b := NewBinaryDecoder([]byte{0, 12, 72, 101, 108, 108, 111, 32, 119, 111, 114, 108, 100, 33})
		var str string
		_ = b.Decode(&str)
		localStr = str
	}
}

func BenchmarkDecode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		localStr = DecodeString(bytes.NewBuffer([]byte{0, 12, 72, 101, 108, 108, 111, 32, 119, 111, 114, 108, 100, 33}))
	}
}

func BenchmarkBinaryEncoder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b := NewBinaryEncoder()
		_ = b.Encode("Hello world!")
		localBytes = b.Bytes()
	}
}

func BenchmarkEncode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buff := bytes.NewBuffer([]byte{})
		EncodeString("Hello world!", buff)
		localBytes = buff.Bytes()
	}
}
