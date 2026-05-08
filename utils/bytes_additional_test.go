package utils

import (
	"bytes"
	"testing"
)

func TestByteHelpers(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "least significant byte",
			run: func(t *testing.T) {
				t.Helper()

				if got := GetByteLSB(0x01020304, 0); got != 0x04 {
					t.Fatalf("byte = %#x, want %#x", got, byte(0x04))
				}

				if got := GetByteLSB(0x01020304, 3); got != 0x01 {
					t.Fatalf("byte = %#x, want %#x", got, byte(0x01))
				}
			},
		},
		{
			name: "bit and nibble helpers",
			run: func(t *testing.T) {
				t.Helper()

				if !GetBitLSB(0b00000100, 2) {
					t.Fatal("bit 2 should be set")
				}

				if GetBitLSB(0b00000100, 1) {
					t.Fatal("bit 1 should not be set")
				}

				if got := LeftNibble(0xAB); got != 0x0A {
					t.Fatalf("left nibble = %#x, want %#x", got, 0x0A)
				}

				if got := RightNibble(0xAB); got != 0x0B {
					t.Fatalf("right nibble = %#x, want %#x", got, 0x0B)
				}
			},
		},
		{
			name: "unsigned and lsb integer helpers",
			run: func(t *testing.T) {
				t.Helper()

				if got := UnsignedByteToInt(0xff); got != 255 {
					t.Fatalf("unsigned = %d, want 255", got)
				}

				if got := LSBBytesToInt([]byte{0x34, 0x12}); got != 0x1234 {
					t.Fatalf("int = %#x, want %#x", got, 0x1234)
				}

				if got := LSBBitValue(3, true); got != 0x08 {
					t.Fatalf("bit value = %#x, want %#x", got, byte(0x08))
				}

				if got := LSBBitValue(3, false); got != 0 {
					t.Fatalf("unset bit value = %#x, want 0", got)
				}
			},
		},
		{
			name: "jam crc32",
			run: func(t *testing.T) {
				t.Helper()

				got := JamCRC32([]byte("123456789"))
				want := []byte{0xd9, 0xc6, 0x0b, 0x34}

				if !bytes.Equal(got, want) {
					t.Fatalf("jam crc = %x, want %x", got, want)
				}
			},
		},
		{
			name: "rotate helpers",
			run: func(t *testing.T) {
				t.Helper()

				left := RotateLeft([]byte{1, 2, 3, 4}, 1)
				if !bytes.Equal(left, []byte{2, 3, 4, 1}) {
					t.Fatalf("left = %v", left)
				}

				right := RotateRight([]byte{1, 2, 3, 4}, 1)
				if !bytes.Equal(right, []byte{4, 1, 2, 3}) {
					t.Fatalf("right = %v", right)
				}
			},
		},
		{
			name: "message padding",
			run: func(t *testing.T) {
				t.Helper()

				padded := PadMessageToBlocksize([]byte{1, 2, 3}, 4)
				if !bytes.Equal(padded, []byte{1, 2, 3, messagePadStart}) {
					t.Fatalf("padded = %v", padded)
				}

				complete := []byte{1, 2, 3, 4}
				if got := PadMessageToBlocksize(complete, 4); !bytes.Equal(got, complete) {
					t.Fatalf("complete = %v, want %v", got, complete)
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}
