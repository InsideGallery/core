package dataconv

import (
	"bytes"
	"math"
	"testing"
)

func TestEncodeBool(t *testing.T) {
	cases := []struct {
		name string
		in   bool
		want byte
	}{
		{"true", true, 1},
		{"false", false, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			EncodeBool(tc.in, buf)
			if buf.Len() != 1 {
				t.Fatalf("expected 1 byte, got %d", buf.Len())
			}
			if buf.Bytes()[0] != tc.want {
				t.Fatalf("expected %d, got %d", tc.want, buf.Bytes()[0])
			}
		})
	}
}

func TestDecodeBool(t *testing.T) {
	cases := []struct {
		name string
		in   []byte
		want bool
	}{
		{"true from 1", []byte{1}, true},
		{"false from 0", []byte{0}, false},
		{"false from 2", []byte{2}, false},
		{"false from 255", []byte{255}, false},
		{"false from 128", []byte{128}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := bytes.NewBuffer(tc.in)
			got := DecodeBool(buf)
			if got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestEncodeDecodeUint8(t *testing.T) {
	cases := []struct {
		name string
		in   uint8
	}{
		{"zero", 0},
		{"one", 1},
		{"mid", 128},
		{"max", 255},
		{"arbitrary", 42},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			EncodeUint8(tc.in, buf)
			got := DecodeUint8(bytes.NewBuffer(buf.Bytes()))
			if got != tc.in {
				t.Fatalf("expected %d, got %d", tc.in, got)
			}
		})
	}
}

func TestEncodeDecodeUint16(t *testing.T) {
	cases := []struct {
		name string
		in   uint16
	}{
		{"zero", 0},
		{"one", 1},
		{"max", math.MaxUint16},
		{"mid", 32768},
		{"arbitrary", 1245},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			EncodeUint16(tc.in, buf)
			if buf.Len() != 2 {
				t.Fatalf("expected 2 bytes, got %d", buf.Len())
			}
			got := DecodeUint16(bytes.NewBuffer(buf.Bytes()))
			if got != tc.in {
				t.Fatalf("expected %d, got %d", tc.in, got)
			}
		})
	}
}

func TestEncodeDecodeUint32(t *testing.T) {
	cases := []struct {
		name string
		in   uint32
	}{
		{"zero", 0},
		{"one", 1},
		{"max", math.MaxUint32},
		{"mid", 2147483648},
		{"arbitrary", 1245678},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			EncodeUint32(tc.in, buf)
			if buf.Len() != 4 {
				t.Fatalf("expected 4 bytes, got %d", buf.Len())
			}
			got := DecodeUint32(bytes.NewBuffer(buf.Bytes()))
			if got != tc.in {
				t.Fatalf("expected %d, got %d", tc.in, got)
			}
		})
	}
}

func TestEncodeDecodeUint64(t *testing.T) {
	cases := []struct {
		name string
		in   uint64
	}{
		{"zero", 0},
		{"one", 1},
		{"max", math.MaxUint64},
		{"mid", 9223372036854775808},
		{"arbitrary", 124567891011},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			EncodeUint64(tc.in, buf)
			if buf.Len() != 8 {
				t.Fatalf("expected 8 bytes, got %d", buf.Len())
			}
			got := DecodeUint64(bytes.NewBuffer(buf.Bytes()))
			if got != tc.in {
				t.Fatalf("expected %d, got %d", tc.in, got)
			}
		})
	}
}

func TestEncodeDecodeFloat64(t *testing.T) {
	cases := []struct {
		name string
		in   float64
	}{
		{"zero", 0.0},
		{"positive", 3.14159265358979},
		{"negative", -273.15},
		{"max", math.MaxFloat64},
		{"smallest_positive", math.SmallestNonzeroFloat64},
		{"one", 1.0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			EncodeFloat64(tc.in, buf)
			if buf.Len() != 8 {
				t.Fatalf("expected 8 bytes, got %d", buf.Len())
			}
			got := DecodeFloat64(bytes.NewBuffer(buf.Bytes()))
			if got != tc.in {
				t.Fatalf("expected %f, got %f", tc.in, got)
			}
		})
	}
}

func TestEncodeDecodeFloat32(t *testing.T) {
	cases := []struct {
		name string
		in   float32
	}{
		{"zero", 0.0},
		{"positive", 3.14},
		{"negative", -100.5},
		{"max", math.MaxFloat32},
		{"smallest_positive", math.SmallestNonzeroFloat32},
		{"one", 1.0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			EncodeFloat32(tc.in, buf)
			if buf.Len() != 4 {
				t.Fatalf("expected 4 bytes, got %d", buf.Len())
			}
			got := DecodeFloat32(bytes.NewBuffer(buf.Bytes()))
			if got != tc.in {
				t.Fatalf("expected %f, got %f", tc.in, got)
			}
		})
	}
}

func TestEncodeDecodeInt(t *testing.T) {
	cases := []struct {
		name string
		in   int
		want int
	}{
		{"zero", 0, 0},
		{"one", 1, 1},
		{"large_positive", 2147483647, 2147483647},
		{"arbitrary", 123456, 123456},
		{"max_uint32_range", 4294967295, 4294967295},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			EncodeInt(tc.in, buf)
			if buf.Len() != 4 {
				t.Fatalf("expected 4 bytes, got %d", buf.Len())
			}
			got := DecodeInt(bytes.NewBuffer(buf.Bytes()))
			if got != tc.want {
				t.Fatalf("expected %d, got %d", tc.want, got)
			}
		})
	}
}

func TestEncodeDecodeInt64(t *testing.T) {
	cases := []struct {
		name string
		in   int64
	}{
		{"zero", 0},
		{"one", 1},
		{"negative_one", -1},
		{"max", math.MaxInt64},
		{"min", math.MinInt64},
		{"arbitrary", 9876543210},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			EncodeInt64(tc.in, buf)
			if buf.Len() != 8 {
				t.Fatalf("expected 8 bytes, got %d", buf.Len())
			}
			got := DecodeInt64(bytes.NewBuffer(buf.Bytes()))
			if got != tc.in {
				t.Fatalf("expected %d, got %d", tc.in, got)
			}
		})
	}
}

func TestEncodeDecodeInt32(t *testing.T) {
	cases := []struct {
		name string
		in   int32
	}{
		{"zero", 0},
		{"one", 1},
		{"negative_one", -1},
		{"max", math.MaxInt32},
		{"min", math.MinInt32},
		{"arbitrary", 42},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			EncodeInt32(tc.in, buf)
			if buf.Len() != 4 {
				t.Fatalf("expected 4 bytes, got %d", buf.Len())
			}
			got := DecodeInt32(bytes.NewBuffer(buf.Bytes()))
			if got != tc.in {
				t.Fatalf("expected %d, got %d", tc.in, got)
			}
		})
	}
}

func TestEncodeDecodeInt16(t *testing.T) {
	cases := []struct {
		name string
		in   int16
	}{
		{"zero", 0},
		{"one", 1},
		{"negative_one", -1},
		{"max", math.MaxInt16},
		{"min", math.MinInt16},
		{"arbitrary", 1234},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			EncodeInt16(tc.in, buf)
			if buf.Len() != 2 {
				t.Fatalf("expected 2 bytes, got %d", buf.Len())
			}
			got := DecodeInt16(bytes.NewBuffer(buf.Bytes()))
			if got != tc.in {
				t.Fatalf("expected %d, got %d", tc.in, got)
			}
		})
	}
}

func TestEncodeDecodeInt8(t *testing.T) {
	cases := []struct {
		name string
		in   int8
	}{
		{"zero", 0},
		{"one", 1},
		{"negative_one", -1},
		{"max", math.MaxInt8},
		{"min", math.MinInt8},
		{"arbitrary", 42},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			EncodeInt8(tc.in, buf)
			if buf.Len() != 1 {
				t.Fatalf("expected 1 byte, got %d", buf.Len())
			}
			got := DecodeInt8(bytes.NewBuffer(buf.Bytes()))
			if got != tc.in {
				t.Fatalf("expected %d, got %d", tc.in, got)
			}
		})
	}
}

func TestEncodeDecodeBytes(t *testing.T) {
	cases := []struct {
		name string
		in   []byte
	}{
		{"empty", []byte{}},
		{"single_byte", []byte{0xFF}},
		{"hello", []byte("hello")},
		{"binary_data", []byte{0, 1, 2, 3, 4, 5}},
		{"null_bytes", []byte{0, 0, 0, 0, 0}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			EncodeBytes(tc.in, buf)
			got := DecodeBytes(bytes.NewBuffer(buf.Bytes()))
			if !bytes.Equal(got, tc.in) {
				t.Fatalf("expected %v, got %v", tc.in, got)
			}
		})
	}
}

func TestEncodeDecodeString(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{"empty", ""},
		{"single_char", "a"},
		{"hello_world", "Hello world!"},
		{"unicode", "\u00e9\u00e0\u00fc\u00f1"},
		{"spaces_only", "   "},
		{"long_string", "abcdefghijklmnopqrstuvwxyz0123456789"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			EncodeString(tc.in, buf)
			got := DecodeString(bytes.NewBuffer(buf.Bytes()))
			if got != tc.in {
				t.Fatalf("expected %q, got %q", tc.in, got)
			}
		})
	}
}

func TestEncodeDecodeBuffer(t *testing.T) {
	cases := []struct {
		name string
		in   []byte
	}{
		{"empty", []byte{}},
		{"single_byte", []byte{42}},
		{"multiple_bytes", []byte{1, 2, 3, 4, 5}},
		{"binary_data", []byte{0xFF, 0xFE, 0xFD}},
		{"large_buffer", bytes.Repeat([]byte{0xAB}, 1024)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srcBuf := bytes.NewBuffer(tc.in)
			dstBuf := bytes.NewBuffer(nil)
			EncodeBuffer(srcBuf, dstBuf)
			got := DecodeBuffer(bytes.NewBuffer(dstBuf.Bytes()))
			if !bytes.Equal(got.Bytes(), tc.in) {
				t.Fatalf("expected %v, got %v", tc.in, got.Bytes())
			}
		})
	}
}

type testEncodeable struct {
	kind uint8
	data []byte
}

func (e *testEncodeable) Encode() (uint8, []byte) {
	return e.kind, e.data
}

func (e *testEncodeable) Decode(kind uint8, data []byte) any {
	return &testEncodeable{kind: kind, data: data}
}

func TestEncodeEncodeable(t *testing.T) {
	cases := []struct {
		name     string
		kind     uint8
		data     []byte
		wantLen  int
	}{
		{"empty_data", 1, []byte{}, 1},
		{"single_byte_data", 2, []byte{0xFF}, 2},
		{"multi_byte_data", 3, []byte{1, 2, 3}, 4},
		{"kind_zero", 0, []byte{10, 20}, 3},
		{"kind_max", 255, []byte{100}, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			enc := &testEncodeable{kind: tc.kind, data: tc.data}
			EncodeEncodeable(enc, buf)
			if buf.Len() != tc.wantLen {
				t.Fatalf("expected %d bytes, got %d", tc.wantLen, buf.Len())
			}
			if buf.Bytes()[0] != tc.kind {
				t.Fatalf("expected kind %d, got %d", tc.kind, buf.Bytes()[0])
			}
		})
	}
}

func TestDecodeDecodable(t *testing.T) {
	cases := []struct {
		name string
		kind uint8
		data []byte
	}{
		{"empty_data", 1, []byte{}},
		{"single_byte", 2, []byte{0xAA}},
		{"multi_byte", 3, []byte{10, 20, 30}},
		{"kind_zero", 0, []byte{5, 6}},
		{"kind_max", 255, []byte{99}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw := append([]byte{tc.kind}, tc.data...)
			buf := bytes.NewBuffer(raw)
			dec := &testEncodeable{}
			result := DecodeDecodable(dec, buf)
			got, ok := result.(*testEncodeable)
			if !ok {
				t.Fatal("expected *testEncodeable")
			}
			if got.kind != tc.kind {
				t.Fatalf("expected kind %d, got %d", tc.kind, got.kind)
			}
			if !bytes.Equal(got.data, tc.data) {
				t.Fatalf("expected data %v, got %v", tc.data, got.data)
			}
		})
	}
}

func TestBinaryEncoderEncodeBool(t *testing.T) {
	cases := []struct {
		name string
		in   bool
		want bool
	}{
		{"true", true, true},
		{"false", false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			enc := NewBinaryEncoder()
			if err := enc.Encode(tc.in); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			dec := NewBinaryDecoder(enc.Bytes())
			var got bool
			if err := dec.Decode(&got); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestBinaryEncoderEncodeInt(t *testing.T) {
	cases := []struct {
		name string
		in   int
		want int
	}{
		{"zero", 0, 0},
		{"positive", 42, 42},
		{"large", 2147483647, 2147483647},
		{"one", 1, 1},
		{"boundary", 256, 256},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			enc := NewBinaryEncoder()
			if err := enc.Encode(tc.in); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			dec := NewBinaryDecoder(enc.Bytes())
			var got int
			if err := dec.Decode(&got); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("expected %d, got %d", tc.want, got)
			}
		})
	}
}

func TestBinaryEncoderEncodeInt8(t *testing.T) {
	cases := []struct {
		name string
		in   int8
	}{
		{"zero", 0},
		{"positive", 42},
		{"negative", -1},
		{"max", math.MaxInt8},
		{"min", math.MinInt8},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			enc := NewBinaryEncoder()
			if err := enc.Encode(tc.in); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			dec := NewBinaryDecoder(enc.Bytes())
			var got int8
			if err := dec.Decode(&got); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.in {
				t.Fatalf("expected %d, got %d", tc.in, got)
			}
		})
	}
}

func TestBinaryEncoderEncodeInt16(t *testing.T) {
	cases := []struct {
		name string
		in   int16
	}{
		{"zero", 0},
		{"positive", 1234},
		{"negative", -1234},
		{"max", math.MaxInt16},
		{"min", math.MinInt16},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			enc := NewBinaryEncoder()
			if err := enc.Encode(tc.in); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			dec := NewBinaryDecoder(enc.Bytes())
			var got int16
			if err := dec.Decode(&got); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.in {
				t.Fatalf("expected %d, got %d", tc.in, got)
			}
		})
	}
}

func TestBinaryEncoderEncodeInt32(t *testing.T) {
	cases := []struct {
		name string
		in   int32
	}{
		{"zero", 0},
		{"positive", 123456},
		{"negative", -123456},
		{"max", math.MaxInt32},
		{"min", math.MinInt32},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			enc := NewBinaryEncoder()
			if err := enc.Encode(tc.in); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			dec := NewBinaryDecoder(enc.Bytes())
			var got int32
			if err := dec.Decode(&got); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.in {
				t.Fatalf("expected %d, got %d", tc.in, got)
			}
		})
	}
}

func TestBinaryEncoderEncodeInt64(t *testing.T) {
	cases := []struct {
		name string
		in   int64
	}{
		{"zero", 0},
		{"positive", 9876543210},
		{"negative", -9876543210},
		{"max", math.MaxInt64},
		{"min", math.MinInt64},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			enc := NewBinaryEncoder()
			if err := enc.Encode(tc.in); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			dec := NewBinaryDecoder(enc.Bytes())
			var got int64
			if err := dec.Decode(&got); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.in {
				t.Fatalf("expected %d, got %d", tc.in, got)
			}
		})
	}
}

func TestBinaryEncoderEncodeBuffer(t *testing.T) {
	cases := []struct {
		name string
		in   []byte
	}{
		{"empty", []byte{}},
		{"single", []byte{1}},
		{"multi", []byte{1, 2, 3, 4, 5}},
		{"binary", []byte{0xFF, 0x00, 0xAB}},
		{"large", bytes.Repeat([]byte{0xCD}, 256)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			enc := NewBinaryEncoder()
			inBuf := bytes.NewBuffer(tc.in)
			if err := enc.Encode(inBuf); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			dec := NewBinaryDecoder(enc.Bytes())
			var got bytes.Buffer
			if err := dec.Decode(&got); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !bytes.Equal(got.Bytes(), tc.in) {
				t.Fatalf("expected %v, got %v", tc.in, got.Bytes())
			}
		})
	}
}

func TestBinaryEncoderEncodeEncodeable(t *testing.T) {
	cases := []struct {
		name string
		kind uint8
		data []byte
	}{
		{"kind_1_empty", 1, []byte{}},
		{"kind_2_data", 2, []byte{10, 20}},
		{"kind_0_data", 0, []byte{0}},
		{"kind_255_data", 255, []byte{1, 2, 3}},
		{"kind_100_single", 100, []byte{42}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			enc := NewBinaryEncoder()
			obj := &testEncodeable{kind: tc.kind, data: tc.data}
			if err := enc.Encode(obj); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			encoded := enc.Bytes()
			if len(encoded) == 0 {
				t.Fatal("expected non-empty encoded bytes")
			}
			if encoded[0] != tc.kind {
				t.Fatalf("expected kind %d at position 0, got %d", tc.kind, encoded[0])
			}
		})
	}
}

func TestBinaryEncoderEncodeUnsupportedType(t *testing.T) {
	cases := []struct {
		name string
		in   any
	}{
		{"nil", nil},
		{"struct", struct{ X int }{1}},
		{"slice_of_int", []int{1, 2, 3}},
		{"map", map[string]int{"a": 1}},
		{"complex128", complex(1, 2)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			enc := NewBinaryEncoder()
			err := enc.Encode(tc.in)
			if err != ErrWrongEncodeType {
				t.Fatalf("expected ErrWrongEncodeType, got %v", err)
			}
		})
	}
}

func TestBinaryDecoderDecodeUnsupportedType(t *testing.T) {
	cases := []struct {
		name string
		in   any
	}{
		{"nil", nil},
		{"int_not_pointer", 42},
		{"string_not_pointer", "hello"},
		{"slice_of_int", []int{1, 2, 3}},
		{"map", map[string]int{"a": 1}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dec := NewBinaryDecoder([]byte{0, 0, 0, 0, 0, 0, 0, 0})
			err := dec.Decode(tc.in)
			if err != ErrWrongDecodeType {
				t.Fatalf("expected ErrWrongDecodeType, got %v", err)
			}
		})
	}
}

func TestBinaryDecoderDecodeDecodable(t *testing.T) {
	cases := []struct {
		name string
		kind uint8
		data []byte
	}{
		{"kind_1_empty", 1, []byte{}},
		{"kind_2_single", 2, []byte{0xAA}},
		{"kind_3_multi", 3, []byte{10, 20, 30}},
		{"kind_0_data", 0, []byte{5, 6}},
		{"kind_255_data", 255, []byte{99}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw := append([]byte{tc.kind}, tc.data...)
			dec := NewBinaryDecoder(raw)
			obj := &testEncodeable{}
			err := dec.Decode(obj)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestNewBinaryEncoderBytes(t *testing.T) {
	cases := []struct {
		name string
	}{
		{"initial_empty"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			enc := NewBinaryEncoder()
			if enc == nil {
				t.Fatal("expected non-nil encoder")
			}
			b := enc.Bytes()
			if len(b) != 0 {
				t.Fatalf("expected empty bytes, got %d bytes", len(b))
			}
		})
	}
}

func TestNewBinaryDecoderBytes(t *testing.T) {
	cases := []struct {
		name string
		in   []byte
		want []byte
	}{
		{"empty", []byte{}, []byte{}},
		{"single_byte", []byte{42}, []byte{42}},
		{"multi_bytes", []byte{1, 2, 3}, []byte{1, 2, 3}},
		{"nil_input", nil, []byte{}},
		{"large_input", bytes.Repeat([]byte{0xAB}, 100), bytes.Repeat([]byte{0xAB}, 100)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dec := NewBinaryDecoder(tc.in)
			if dec == nil {
				t.Fatal("expected non-nil decoder")
			}
			got := dec.Bytes()
			if !bytes.Equal(got, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestBinaryEncoderMultipleValues(t *testing.T) {
	cases := []struct {
		name string
	}{
		{"bool_int8_string"},
		{"all_int_types"},
		{"mixed_types"},
		{"floats_and_bytes"},
		{"string_and_bool"},
	}

	t.Run(cases[0].name, func(t *testing.T) {
		enc := NewBinaryEncoder()
		if err := enc.Encode(true); err != nil {
			t.Fatal(err)
		}
		if err := enc.Encode(int8(42)); err != nil {
			t.Fatal(err)
		}
		if err := enc.Encode("test"); err != nil {
			t.Fatal(err)
		}
		dec := NewBinaryDecoder(enc.Bytes())
		var b bool
		var i int8
		var s string
		if err := dec.Decode(&b); err != nil {
			t.Fatal(err)
		}
		if err := dec.Decode(&i); err != nil {
			t.Fatal(err)
		}
		if err := dec.Decode(&s); err != nil {
			t.Fatal(err)
		}
		if b != true || i != 42 || s != "test" {
			t.Fatalf("decoded values mismatch: %v %d %q", b, i, s)
		}
	})

	t.Run(cases[1].name, func(t *testing.T) {
		enc := NewBinaryEncoder()
		if err := enc.Encode(int(100)); err != nil {
			t.Fatal(err)
		}
		if err := enc.Encode(int16(200)); err != nil {
			t.Fatal(err)
		}
		if err := enc.Encode(int32(300)); err != nil {
			t.Fatal(err)
		}
		if err := enc.Encode(int64(400)); err != nil {
			t.Fatal(err)
		}
		dec := NewBinaryDecoder(enc.Bytes())
		var i int
		var i16 int16
		var i32 int32
		var i64 int64
		if err := dec.Decode(&i); err != nil {
			t.Fatal(err)
		}
		if err := dec.Decode(&i16); err != nil {
			t.Fatal(err)
		}
		if err := dec.Decode(&i32); err != nil {
			t.Fatal(err)
		}
		if err := dec.Decode(&i64); err != nil {
			t.Fatal(err)
		}
		if i != 100 || i16 != 200 || i32 != 300 || i64 != 400 {
			t.Fatalf("decoded values mismatch: %d %d %d %d", i, i16, i32, i64)
		}
	})

	t.Run(cases[2].name, func(t *testing.T) {
		enc := NewBinaryEncoder()
		if err := enc.Encode(uint8(10)); err != nil {
			t.Fatal(err)
		}
		if err := enc.Encode(float64(3.14)); err != nil {
			t.Fatal(err)
		}
		if err := enc.Encode(false); err != nil {
			t.Fatal(err)
		}
		dec := NewBinaryDecoder(enc.Bytes())
		var u8 uint8
		var f64 float64
		var b bool
		if err := dec.Decode(&u8); err != nil {
			t.Fatal(err)
		}
		if err := dec.Decode(&f64); err != nil {
			t.Fatal(err)
		}
		if err := dec.Decode(&b); err != nil {
			t.Fatal(err)
		}
		if u8 != 10 || f64 != 3.14 || b != false {
			t.Fatalf("decoded values mismatch: %d %f %v", u8, f64, b)
		}
	})

	t.Run(cases[3].name, func(t *testing.T) {
		enc := NewBinaryEncoder()
		if err := enc.Encode(float32(1.5)); err != nil {
			t.Fatal(err)
		}
		if err := enc.Encode(float64(2.5)); err != nil {
			t.Fatal(err)
		}
		if err := enc.Encode([]byte{0xDE, 0xAD}); err != nil {
			t.Fatal(err)
		}
		dec := NewBinaryDecoder(enc.Bytes())
		var f32 float32
		var f64 float64
		var bs []byte
		if err := dec.Decode(&f32); err != nil {
			t.Fatal(err)
		}
		if err := dec.Decode(&f64); err != nil {
			t.Fatal(err)
		}
		if err := dec.Decode(&bs); err != nil {
			t.Fatal(err)
		}
		if f32 != 1.5 || f64 != 2.5 || !bytes.Equal(bs, []byte{0xDE, 0xAD}) {
			t.Fatalf("decoded values mismatch: %f %f %v", f32, f64, bs)
		}
	})

	t.Run(cases[4].name, func(t *testing.T) {
		enc := NewBinaryEncoder()
		if err := enc.Encode("hello"); err != nil {
			t.Fatal(err)
		}
		if err := enc.Encode(true); err != nil {
			t.Fatal(err)
		}
		if err := enc.Encode("world"); err != nil {
			t.Fatal(err)
		}
		dec := NewBinaryDecoder(enc.Bytes())
		var s1, s2 string
		var b bool
		if err := dec.Decode(&s1); err != nil {
			t.Fatal(err)
		}
		if err := dec.Decode(&b); err != nil {
			t.Fatal(err)
		}
		if err := dec.Decode(&s2); err != nil {
			t.Fatal(err)
		}
		if s1 != "hello" || b != true || s2 != "world" {
			t.Fatalf("decoded values mismatch: %q %v %q", s1, b, s2)
		}
	})
}

func TestEncodeDecodeBoolRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   bool
	}{
		{"true_via_encoder", true},
		{"false_via_encoder", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			EncodeBool(tc.in, buf)
			got := DecodeBool(bytes.NewBuffer(buf.Bytes()))
			if got != tc.in {
				t.Fatalf("expected %v, got %v", tc.in, got)
			}
		})
	}
}
