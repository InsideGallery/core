package dataconv

import (
	"bytes"
	"encoding/binary"
	"math"

	"github.com/InsideGallery/core/utils"
)

// Encodeable describe encodeable interface
type Encodeable interface {
	Encode() (uint8, []byte)
}

// Decodable describe decodable interface
type Decodable interface {
	Decode(uint8, []byte) any
}

// BinaryEncoder contains buffer and convert values to binary
type BinaryEncoder struct {
	data *bytes.Buffer
}

// NewBinaryEncoder return new encoder
func NewBinaryEncoder() *BinaryEncoder {
	return &BinaryEncoder{
		data: bytes.NewBuffer([]byte{}),
	}
}

// EncodeBool encode bool
func EncodeBool(v bool, buffer *bytes.Buffer) {
	if v {
		buffer.Write([]byte{1})
	} else {
		buffer.Write([]byte{0})
	}
}

// EncodeUint8 encode uint8
func EncodeUint8(v uint8, buffer *bytes.Buffer) {
	buffer.Write([]byte{v})
}

// EncodeUint16 encode uint16
func EncodeUint16(v uint16, buffer *bytes.Buffer) {
	data := make([]byte, 2) //nolint:mnd
	binary.BigEndian.PutUint16(data, v)
	buffer.Write(data)
}

// EncodeUint32 encode uint32
func EncodeUint32(v uint32, buffer *bytes.Buffer) {
	data := make([]byte, 4) //nolint:mnd
	binary.BigEndian.PutUint32(data, v)
	buffer.Write(data)
}

// EncodeUint64 encode uint64
func EncodeUint64(v uint64, buffer *bytes.Buffer) {
	data := make([]byte, 8) //nolint:mnd
	binary.BigEndian.PutUint64(data, v)
	buffer.Write(data)
}

// EncodeFloat64 encode float64
func EncodeFloat64(v float64, buffer *bytes.Buffer) {
	data := make([]byte, 8) //nolint:mnd
	binary.BigEndian.PutUint64(data, math.Float64bits(v))
	buffer.Write(data)
}

// EncodeFloat32 encode float32
func EncodeFloat32(v float32, buffer *bytes.Buffer) {
	data := make([]byte, 4) //nolint:mnd
	binary.BigEndian.PutUint32(data, math.Float32bits(v))
	buffer.Write(data)
}

// EncodeInt encode int
func EncodeInt(v int, buffer *bytes.Buffer) {
	data := make([]byte, 4) //nolint:mnd
	binary.BigEndian.PutUint32(data, uint32(v))
	buffer.Write(data)
}

// EncodeInt64 encode int64
func EncodeInt64(v int64, buffer *bytes.Buffer) {
	data := make([]byte, 8)                     //nolint:mnd
	binary.BigEndian.PutUint64(data, uint64(v)) //nolint:gosec
	buffer.Write(data)
}

// EncodeInt32 encode int32
func EncodeInt32(v int32, buffer *bytes.Buffer) {
	data := make([]byte, 4)                     //nolint:mnd
	binary.BigEndian.PutUint32(data, uint32(v)) //nolint:gosec
	buffer.Write(data)
}

// EncodeInt16 encode int16
func EncodeInt16(v int16, buffer *bytes.Buffer) {
	data := make([]byte, 2)                     //nolint:mnd
	binary.BigEndian.PutUint16(data, uint16(v)) //nolint:gosec
	buffer.Write(data)
}

// EncodeInt8 encode int8
func EncodeInt8(v int8, buffer *bytes.Buffer) {
	buffer.Write([]byte{uint8(v)}) //nolint:gosec
}

// EncodeBytes encode bytes
func EncodeBytes(v []byte, buffer *bytes.Buffer) {
	data := make([]byte, 2) //nolint:mnd
	binary.BigEndian.PutUint16(data, uint16(len(v)))
	buffer.Write(data)
	buffer.Write(v)
}

// EncodeString encode string
func EncodeString(v string, buffer *bytes.Buffer) {
	rawString := []byte(v)
	data := make([]byte, 2) //nolint:mnd
	binary.BigEndian.PutUint16(data, uint16(len(rawString)))
	buffer.Write(data)
	buffer.Write(rawString)
}

// EncodeBuffer encode buffer
func EncodeBuffer(v *bytes.Buffer, buffer *bytes.Buffer) {
	buffer.Write(v.Bytes())
}

// EncodeEncodeable encode Encodeable
func EncodeEncodeable(v Encodeable, buffer *bytes.Buffer) {
	kind, data := v.Encode()
	buffer.Write([]byte{kind})
	buffer.Write(data)
}

// Encode add value to bytes
func (b *BinaryEncoder) Encode(value any) error {
	switch v := value.(type) {
	case bool:
		EncodeBool(v, b.data)
	case uint8:
		EncodeUint8(v, b.data)
	case uint16:
		EncodeUint16(v, b.data)
	case uint32:
		EncodeUint32(v, b.data)
	case uint64:
		EncodeUint64(v, b.data)
	case float64:
		EncodeFloat64(v, b.data)
	case float32:
		EncodeFloat32(v, b.data)
	case int:
		EncodeInt(v, b.data)
	case int64:
		EncodeInt64(v, b.data)
	case int32:
		EncodeInt32(v, b.data)
	case int16:
		EncodeInt16(v, b.data)
	case int8:
		EncodeInt8(v, b.data)
	case []byte:
		EncodeBytes(v, b.data)
	case string:
		EncodeString(v, b.data)
	case *bytes.Buffer:
		EncodeBuffer(v, b.data)
	case Encodeable:
		EncodeEncodeable(v, b.data)
	default:
		return ErrWrongEncodeType
	}

	return nil
}

// Bytes return bytes
func (b *BinaryEncoder) Bytes() []byte {
	return b.data.Bytes()
}

// BinaryDecoder contains buffer and convert binary to value
type BinaryDecoder struct {
	data *bytes.Buffer
}

// NewBinaryDecoder return new decoder
func NewBinaryDecoder(data []byte) *BinaryDecoder {
	return &BinaryDecoder{
		data: bytes.NewBuffer(data),
	}
}

// DecodeBool decode to bool
func DecodeBool(buffer *bytes.Buffer) bool {
	data := buffer.Next(1) //nolint:mnd
	return data[0] == 1    //nolint:mnd
}

// DecodeUint8 decode to uint8
func DecodeUint8(buffer *bytes.Buffer) uint8 {
	data := buffer.Next(1) //nolint:mnd
	return data[0]
}

// DecodeUint16 decode to uint16
func DecodeUint16(buffer *bytes.Buffer) uint16 {
	return binary.BigEndian.Uint16(buffer.Next(2)) //nolint:mnd
}

// DecodeUint32 decode to uint32
func DecodeUint32(buffer *bytes.Buffer) uint32 {
	return binary.BigEndian.Uint32(buffer.Next(4)) //nolint:mnd
}

// DecodeUint64 decode to uint64
func DecodeUint64(buffer *bytes.Buffer) uint64 {
	return binary.BigEndian.Uint64(buffer.Next(8)) //nolint:mnd
}

// DecodeFloat64 decode to float64
func DecodeFloat64(buffer *bytes.Buffer) float64 {
	bits := binary.BigEndian.Uint64(buffer.Next(8)) //nolint:mnd
	return math.Float64frombits(bits)
}

// DecodeFloat32 decode to float32
func DecodeFloat32(buffer *bytes.Buffer) float32 {
	bits := binary.BigEndian.Uint32(buffer.Next(4)) //nolint:mnd
	return math.Float32frombits(bits)
}

// DecodeInt decode to int
func DecodeInt(buffer *bytes.Buffer) int {
	return int(binary.BigEndian.Uint32(buffer.Next(4))) //nolint:mnd
}

// DecodeInt32 decode to int32
func DecodeInt32(buffer *bytes.Buffer) int32 {
	return int32(binary.BigEndian.Uint32(buffer.Next(4))) //nolint:mnd
}

// DecodeInt64 decode to int64
func DecodeInt64(buffer *bytes.Buffer) int64 {
	return int64(binary.BigEndian.Uint64(buffer.Next(8))) //nolint:mnd
}

// DecodeInt16 decode to int16
func DecodeInt16(buffer *bytes.Buffer) int16 {
	return int16(binary.BigEndian.Uint16(buffer.Next(2))) //nolint:mnd
}

// DecodeInt8 decode to int8
func DecodeInt8(buffer *bytes.Buffer) int8 {
	data := buffer.Next(1) //nolint:mnd
	return int8(data[0])
}

// DecodeBytes decode to bytes
func DecodeBytes(buffer *bytes.Buffer) []byte {
	l := binary.BigEndian.Uint16(buffer.Next(2)) //nolint:mnd
	return buffer.Next(int(l))
}

// DecodeString decode to string
func DecodeString(buffer *bytes.Buffer) string {
	l := binary.BigEndian.Uint16(buffer.Next(2)) //nolint:mnd
	return utils.ByteSliceToString(buffer.Next(int(l)))
}

// DecodeDecodable decode specific structure
func DecodeDecodable(decodable Decodable, buffer *bytes.Buffer) any {
	kind := DecodeUint8(buffer)
	return decodable.Decode(kind, buffer.Next(buffer.Len()))
}

// DecodeBuffer decode to buffer
func DecodeBuffer(buffer *bytes.Buffer) *bytes.Buffer {
	return bytes.NewBuffer(buffer.Bytes())
}

// Decode read value to bytes
func (b *BinaryDecoder) Decode(value any) error {
	switch v := value.(type) {
	case *bool:
		res := DecodeBool(b.data)
		*v = res
	case *uint8:
		res := DecodeUint8(b.data)
		*v = res
	case *uint16:
		res := DecodeUint16(b.data)
		*v = res
	case *uint32:
		res := DecodeUint32(b.data)
		*v = res
	case *uint64:
		res := DecodeUint64(b.data)
		*v = res
	case *float64:
		res := DecodeFloat64(b.data)
		*v = res
	case *float32:
		res := DecodeFloat32(b.data)
		*v = res
	case *int:
		res := DecodeInt(b.data)
		*v = res
	case *int64:
		res := DecodeInt64(b.data)
		*v = res
	case *int32:
		res := DecodeInt32(b.data)
		*v = res
	case *int16:
		res := DecodeInt16(b.data)
		*v = res
	case *int8:
		res := DecodeInt8(b.data)
		*v = res
	case *[]byte:
		res := DecodeBytes(b.data)
		*v = res
	case *string:
		res := DecodeString(b.data)
		*v = res
	case *bytes.Buffer:
		res := DecodeBuffer(b.data)
		*v = *res
	case Decodable:
		DecodeDecodable(v, b.data)
	default:
		return ErrWrongDecodeType
	}

	return nil
}

// Bytes return bytes
func (b *BinaryDecoder) Bytes() []byte {
	return b.data.Bytes()
}
