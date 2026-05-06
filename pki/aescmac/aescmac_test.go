package aescmac

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestXor(t *testing.T) {
	cases := []struct {
		name     string
		a        []byte
		b        []byte
		expected []byte
	}{
		{
			name:     "empty slices",
			a:        []byte{},
			b:        []byte{},
			expected: []byte{},
		},
		{
			name:     "single byte zeros",
			a:        []byte{0x00},
			b:        []byte{0x00},
			expected: []byte{0x00},
		},
		{
			name:     "single byte xor",
			a:        []byte{0xFF},
			b:        []byte{0x0F},
			expected: []byte{0xF0},
		},
		{
			name:     "length mismatch returns nil",
			a:        []byte{0x01, 0x02},
			b:        []byte{0x01},
			expected: nil,
		},
		{
			name:     "identical bytes produce zeros",
			a:        []byte{0xAB, 0xCD, 0xEF},
			b:        []byte{0xAB, 0xCD, 0xEF},
			expected: []byte{0x00, 0x00, 0x00},
		},
		{
			name:     "block size 16 bytes",
			a:        []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10},
			b:        []byte{0x10, 0x0F, 0x0E, 0x0D, 0x0C, 0x0B, 0x0A, 0x09, 0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01},
			expected: []byte{0x11, 0x0D, 0x0D, 0x09, 0x09, 0x0D, 0x0D, 0x01, 0x01, 0x0D, 0x0D, 0x09, 0x09, 0x0D, 0x0D, 0x11},
		},
		{
			name:     "xor with all ones is bitwise not",
			a:        []byte{0x00, 0xFF, 0xAA, 0x55},
			b:        []byte{0xFF, 0xFF, 0xFF, 0xFF},
			expected: []byte{0xFF, 0x00, 0x55, 0xAA},
		},
		{
			name:     "a nil b non-nil returns nil",
			a:        nil,
			b:        []byte{0x01},
			expected: nil,
		},
		{
			name:     "both nil returns empty",
			a:        nil,
			b:        nil,
			expected: []byte{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := Xor(tc.a, tc.b)
			if !bytes.Equal(result, tc.expected) {
				t.Fatalf("Xor(%x, %x) = %x, want %x", tc.a, tc.b, result, tc.expected)
			}
		})
	}
}

func TestShiftLeft(t *testing.T) {
	cases := []struct {
		name     string
		data     []byte
		expected []byte
	}{
		{
			name:     "empty slice",
			data:     []byte{},
			expected: []byte{},
		},
		{
			name:     "single byte zero",
			data:     []byte{0x00},
			expected: []byte{0x00},
		},
		{
			name:     "single byte no carry",
			data:     []byte{0x01},
			expected: []byte{0x02},
		},
		{
			name:     "single byte with MSB set loses top bit",
			data:     []byte{0x80},
			expected: []byte{0x00},
		},
		{
			name:     "single byte 0xFF",
			data:     []byte{0xFF},
			expected: []byte{0xFE},
		},
		{
			name:     "carry propagation across bytes",
			data:     []byte{0x00, 0x80},
			expected: []byte{0x01, 0x00},
		},
		{
			name:     "no carry across bytes",
			data:     []byte{0x01, 0x01},
			expected: []byte{0x02, 0x02},
		},
		{
			name:     "full 16 byte block shift",
			data:     []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
			expected: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02},
		},
		{
			name:     "all FF bytes",
			data:     []byte{0xFF, 0xFF, 0xFF, 0xFF},
			expected: []byte{0xFF, 0xFF, 0xFF, 0xFE},
		},
		{
			name:     "alternating bits 0xAA",
			data:     []byte{0xAA, 0xAA},
			expected: []byte{0x55, 0x54},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := ShiftLeft(tc.data)
			if !bytes.Equal(result, tc.expected) {
				t.Fatalf("ShiftLeft(%x) = %x, want %x", tc.data, result, tc.expected)
			}
		})
	}
}

func TestPadding(t *testing.T) {
	cases := []struct {
		name     string
		data     []byte
		expected []byte
	}{
		{
			name:     "empty data pads to full block",
			data:     []byte{},
			expected: []byte{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:     "single byte pads to block",
			data:     []byte{0x01},
			expected: []byte{0x01, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:     "14 bytes pads to 16",
			data:     []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E},
			expected: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x80, 0x00},
		},
		{
			name:     "15 bytes gets just padding marker",
			data:     []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F},
			expected: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x80},
		},
		{
			name:     "full block gets only padding byte appended",
			data:     []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10},
			expected: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x80},
		},
		{
			name:     "all zeros single byte",
			data:     []byte{0x00},
			expected: []byte{0x00, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:     "7 bytes pads with 8 zeros plus marker",
			data:     []byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF, 0x11},
			expected: []byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF, 0x11, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := Padding(tc.data)
			if !bytes.Equal(result, tc.expected) {
				t.Fatalf("Padding(%x) = %x, want %x", tc.data, result, tc.expected)
			}
		})
	}
}

func TestPaddingLength(t *testing.T) {
	cases := []struct {
		name        string
		inputLen    int
		expectedLen int
	}{
		{
			name:        "empty input produces blockSize",
			inputLen:    0,
			expectedLen: 16,
		},
		{
			name:        "1 byte produces blockSize",
			inputLen:    1,
			expectedLen: 16,
		},
		{
			name:        "15 bytes produces blockSize",
			inputLen:    15,
			expectedLen: 16,
		},
		{
			name:        "16 bytes produces blockSize plus 1",
			inputLen:    16,
			expectedLen: 17,
		},
		{
			name:        "17 bytes produces 18",
			inputLen:    17,
			expectedLen: 18,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := make([]byte, tc.inputLen)
			result := Padding(data)
			if len(result) != tc.expectedLen {
				t.Fatalf("Padding(len=%d) produced len=%d, want %d", tc.inputLen, len(result), tc.expectedLen)
			}
		})
	}
}

func TestNewCMAC(t *testing.T) {
	cases := []struct {
		name      string
		keySize   int
		expectErr error
	}{
		{
			name:      "valid 16 byte key",
			keySize:   16,
			expectErr: nil,
		},
		{
			name:      "valid 24 byte key",
			keySize:   24,
			expectErr: nil,
		},
		{
			name:      "valid 32 byte key",
			keySize:   32,
			expectErr: nil,
		},
		{
			name:      "invalid empty key",
			keySize:   0,
			expectErr: ErrUnsupportedKeySize,
		},
		{
			name:      "invalid 1 byte key",
			keySize:   1,
			expectErr: ErrUnsupportedKeySize,
		},
		{
			name:      "invalid 15 byte key",
			keySize:   15,
			expectErr: ErrUnsupportedKeySize,
		},
		{
			name:      "invalid 17 byte key",
			keySize:   17,
			expectErr: ErrUnsupportedKeySize,
		},
		{
			name:      "invalid 31 byte key",
			keySize:   31,
			expectErr: ErrUnsupportedKeySize,
		},
		{
			name:      "invalid 33 byte key",
			keySize:   33,
			expectErr: ErrUnsupportedKeySize,
		},
		{
			name:      "invalid 8 byte key",
			keySize:   8,
			expectErr: ErrUnsupportedKeySize,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key := make([]byte, tc.keySize)
			h, err := NewCMAC(key)
			if tc.expectErr != nil {
				if err != tc.expectErr {
					t.Fatalf("NewCMAC(keySize=%d) error = %v, want %v", tc.keySize, err, tc.expectErr)
				}
				if h != nil {
					t.Fatalf("NewCMAC(keySize=%d) returned non-nil hash on error", tc.keySize)
				}
			} else {
				if err != nil {
					t.Fatalf("NewCMAC(keySize=%d) unexpected error: %v", tc.keySize, err)
				}
				if h == nil {
					t.Fatalf("NewCMAC(keySize=%d) returned nil hash", tc.keySize)
				}
			}
		})
	}
}

func TestSizeAndBlockSize(t *testing.T) {
	cases := []struct {
		name    string
		keySize int
	}{
		{
			name:    "16 byte key",
			keySize: 16,
		},
		{
			name:    "24 byte key",
			keySize: 24,
		},
		{
			name:    "32 byte key",
			keySize: 32,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name+"_Size", func(t *testing.T) {
			key := make([]byte, tc.keySize)
			h, err := NewCMAC(key)
			if err != nil {
				t.Fatalf("NewCMAC error: %v", err)
			}
			if h.Size() != 16 {
				t.Fatalf("Size() = %d, want 16", h.Size())
			}
		})
		t.Run(tc.name+"_BlockSize", func(t *testing.T) {
			key := make([]byte, tc.keySize)
			h, err := NewCMAC(key)
			if err != nil {
				t.Fatalf("NewCMAC error: %v", err)
			}
			if h.BlockSize() != 16 {
				t.Fatalf("BlockSize() = %d, want 16", h.BlockSize())
			}
		})
	}
}

func TestWrite(t *testing.T) {
	key := make([]byte, 16)

	cases := []struct {
		name       string
		writes     [][]byte
		expectN    []int
		expectErr  bool
		expectMac  string
	}{
		{
			name:      "empty write returns 0",
			writes:    [][]byte{{}},
			expectN:   []int{0},
			expectErr: false,
			expectMac: "",
		},
		{
			name:      "single byte write",
			writes:    [][]byte{{0x01}},
			expectN:   []int{1},
			expectErr: false,
			expectMac: "",
		},
		{
			name:      "exactly one block",
			writes:    [][]byte{make([]byte, 16)},
			expectN:   []int{16},
			expectErr: false,
			expectMac: "",
		},
		{
			name:      "two blocks at once",
			writes:    [][]byte{make([]byte, 32)},
			expectN:   []int{32},
			expectErr: false,
			expectMac: "",
		},
		{
			name:      "multi-block three blocks",
			writes:    [][]byte{make([]byte, 48)},
			expectN:   []int{48},
			expectErr: false,
			expectMac: "",
		},
		{
			name:      "multiple small writes",
			writes:    [][]byte{{0x01}, {0x02}, {0x03}},
			expectN:   []int{1, 1, 1},
			expectErr: false,
			expectMac: "",
		},
		{
			name:      "write after write accumulates",
			writes:    [][]byte{make([]byte, 15), {0xFF}},
			expectN:   []int{15, 1},
			expectErr: false,
			expectMac: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, err := NewCMAC(key)
			if err != nil {
				t.Fatalf("NewCMAC error: %v", err)
			}
			for i, data := range tc.writes {
				n, werr := h.Write(data)
				if werr != nil && !tc.expectErr {
					t.Fatalf("Write #%d unexpected error: %v", i, werr)
				}
				if n != tc.expectN[i] {
					t.Fatalf("Write #%d returned n=%d, want %d", i, n, tc.expectN[i])
				}
			}
		})
	}
}

func TestWriteAfterSum(t *testing.T) {
	cases := []struct {
		name        string
		initialData []byte
	}{
		{
			name:        "write after sum with empty initial data",
			initialData: []byte{},
		},
		{
			name:        "write after sum with single byte",
			initialData: []byte{0x01},
		},
		{
			name:        "write after sum with full block",
			initialData: make([]byte, 16),
		},
		{
			name:        "write after sum with two blocks",
			initialData: make([]byte, 32),
		},
		{
			name:        "write after sum with partial block",
			initialData: make([]byte, 10),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key := make([]byte, 16)
			h, err := NewCMAC(key)
			if err != nil {
				t.Fatalf("NewCMAC error: %v", err)
			}
			if len(tc.initialData) > 0 {
				_, err = h.Write(tc.initialData)
				if err != nil {
					t.Fatalf("Write error: %v", err)
				}
			}
			_ = h.Sum(nil)

			_, err = h.Write([]byte{0x01})
			if err != nil {
				t.Fatalf("Write after Sum returned unexpected error: %v", err)
			}
		})
	}
}

func TestFinishedFlagOnDirectCmac(t *testing.T) {
	cases := []struct {
		name        string
		initialData []byte
	}{
		{
			name:        "finished flag with empty data",
			initialData: []byte{},
		},
		{
			name:        "finished flag with single byte",
			initialData: []byte{0x01},
		},
		{
			name:        "finished flag with block data",
			initialData: make([]byte, 16),
		},
		{
			name:        "finished flag with multi block",
			initialData: make([]byte, 32),
		},
		{
			name:        "finished flag with partial block",
			initialData: make([]byte, 10),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key := make([]byte, 16)
			h, err := NewCMAC(key)
			if err != nil {
				t.Fatalf("NewCMAC error: %v", err)
			}
			c := h.(*cmac)
			if len(tc.initialData) > 0 {
				_, err = c.Write(tc.initialData)
				if err != nil {
					t.Fatalf("Write error: %v", err)
				}
			}
			if c.finished {
				t.Fatalf("finished should be false before Sum")
			}

			_ = c.Sum(nil)

			if c.finished {
				t.Fatalf("Sum uses value receiver so finished should remain false on original")
			}
		})
	}
}

func TestReset(t *testing.T) {
	cases := []struct {
		name        string
		initialData []byte
		callSum     bool
	}{
		{
			name:        "reset without data",
			initialData: nil,
			callSum:     false,
		},
		{
			name:        "reset after write",
			initialData: []byte{0x01, 0x02, 0x03},
			callSum:     false,
		},
		{
			name:        "reset after sum",
			initialData: []byte{0x01, 0x02, 0x03},
			callSum:     true,
		},
		{
			name:        "reset after sum with full block",
			initialData: make([]byte, 16),
			callSum:     true,
		},
		{
			name:        "reset after sum with multi block",
			initialData: make([]byte, 48),
			callSum:     true,
		},
		{
			name:        "reset after empty write and sum",
			initialData: []byte{},
			callSum:     true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key := make([]byte, 16)
			h, err := NewCMAC(key)
			if err != nil {
				t.Fatalf("NewCMAC error: %v", err)
			}

			if tc.initialData != nil && len(tc.initialData) > 0 {
				_, err = h.Write(tc.initialData)
				if err != nil {
					t.Fatalf("Write error: %v", err)
				}
			}

			if tc.callSum {
				_ = h.Sum(nil)
			}

			h.Reset()

			n, err := h.Write([]byte{0xAB, 0xCD})
			if err != nil {
				t.Fatalf("Write after Reset returned error: %v", err)
			}
			if n != 2 {
				t.Fatalf("Write after Reset returned n=%d, want 2", n)
			}
		})
	}
}

func TestResetProducesSameResult(t *testing.T) {
	cases := []struct {
		name string
		data []byte
	}{
		{
			name: "single byte data",
			data: []byte{0x42},
		},
		{
			name: "partial block",
			data: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
		},
		{
			name: "exact block",
			data: make([]byte, 16),
		},
		{
			name: "multi block",
			data: make([]byte, 33),
		},
		{
			name: "two blocks",
			data: make([]byte, 32),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key := make([]byte, 16)
			h, err := NewCMAC(key)
			if err != nil {
				t.Fatalf("NewCMAC error: %v", err)
			}

			_, _ = h.Write(tc.data)
			mac1 := h.Sum(nil)

			h.Reset()
			_, _ = h.Write(tc.data)
			mac2 := h.Sum(nil)

			if !bytes.Equal(mac1, mac2) {
				t.Fatalf("Reset did not restore state: first=%x, second=%x", mac1, mac2)
			}
		})
	}
}

func TestSumPackageLevel(t *testing.T) {
	cases := []struct {
		name      string
		keySize   int
		data      []byte
		expectErr bool
	}{
		{
			name:      "valid key empty data",
			keySize:   16,
			data:      []byte{},
			expectErr: false,
		},
		{
			name:      "valid key single byte",
			keySize:   16,
			data:      []byte{0x01},
			expectErr: false,
		},
		{
			name:      "valid key full block",
			keySize:   16,
			data:      make([]byte, 16),
			expectErr: false,
		},
		{
			name:      "valid key multi block",
			keySize:   16,
			data:      make([]byte, 64),
			expectErr: false,
		},
		{
			name:      "valid 24 byte key",
			keySize:   24,
			data:      []byte{0x01, 0x02},
			expectErr: false,
		},
		{
			name:      "valid 32 byte key",
			keySize:   32,
			data:      []byte{0x01, 0x02, 0x03},
			expectErr: false,
		},
		{
			name:      "invalid key size returns error",
			keySize:   10,
			data:      []byte{0x01},
			expectErr: true,
		},
		{
			name:      "nil data with valid key",
			keySize:   16,
			data:      nil,
			expectErr: false,
		},
		{
			name:      "empty key returns error",
			keySize:   0,
			data:      []byte{0x01},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key := make([]byte, tc.keySize)
			result, err := Sum(key, tc.data)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("Sum expected error but got nil")
				}
				if result != nil {
					t.Fatalf("Sum returned non-nil result on error")
				}
			} else {
				if err != nil {
					t.Fatalf("Sum unexpected error: %v", err)
				}
				if len(result) != 16 {
					t.Fatalf("Sum result length = %d, want 16", len(result))
				}
			}
		})
	}
}

func TestSumConsistency(t *testing.T) {
	cases := []struct {
		name string
		data []byte
	}{
		{
			name: "empty data",
			data: []byte{},
		},
		{
			name: "single byte",
			data: []byte{0x42},
		},
		{
			name: "partial block",
			data: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07},
		},
		{
			name: "exact block",
			data: make([]byte, 16),
		},
		{
			name: "multi block",
			data: make([]byte, 48),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key := make([]byte, 16)

			result1, err := Sum(key, tc.data)
			if err != nil {
				t.Fatalf("Sum first call error: %v", err)
			}
			result2, err := Sum(key, tc.data)
			if err != nil {
				t.Fatalf("Sum second call error: %v", err)
			}
			if !bytes.Equal(result1, result2) {
				t.Fatalf("Sum not deterministic: %x != %x", result1, result2)
			}
		})
	}
}

func TestSumMethodWithPrefix(t *testing.T) {
	cases := []struct {
		name   string
		prefix []byte
		data   []byte
	}{
		{
			name:   "nil prefix",
			prefix: nil,
			data:   []byte{0x01},
		},
		{
			name:   "empty prefix",
			prefix: []byte{},
			data:   []byte{0x01},
		},
		{
			name:   "non-empty prefix",
			prefix: []byte("test string"),
			data:   []byte{0x01},
		},
		{
			name:   "prefix with full block data",
			prefix: []byte{0xAA, 0xBB},
			data:   make([]byte, 16),
		},
		{
			name:   "long prefix",
			prefix: make([]byte, 100),
			data:   []byte{0x01, 0x02, 0x03},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key := make([]byte, 16)
			h, err := NewCMAC(key)
			if err != nil {
				t.Fatalf("NewCMAC error: %v", err)
			}
			_, _ = h.Write(tc.data)
			result := h.Sum(tc.prefix)

			if len(result) != len(tc.prefix)+16 {
				t.Fatalf("Sum(prefix) length = %d, want %d", len(result), len(tc.prefix)+16)
			}

			if tc.prefix != nil && len(tc.prefix) > 0 {
				if !bytes.Equal(result[:len(tc.prefix)], tc.prefix) {
					t.Fatalf("Sum(prefix) prefix mismatch")
				}
			}
		})
	}
}

func TestSumWithNoWrite(t *testing.T) {
	cases := []struct {
		name    string
		keySize int
	}{
		{
			name:    "16 byte key no write",
			keySize: 16,
		},
		{
			name:    "24 byte key no write",
			keySize: 24,
		},
		{
			name:    "32 byte key no write",
			keySize: 32,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key := make([]byte, tc.keySize)
			h, err := NewCMAC(key)
			if err != nil {
				t.Fatalf("NewCMAC error: %v", err)
			}
			result := h.Sum(nil)
			if len(result) != 16 {
				t.Fatalf("Sum without Write produced len=%d, want 16", len(result))
			}
		})
	}
}

func TestSumKnownVectors(t *testing.T) {
	cases := []struct {
		name        string
		keyHex      string
		dataHex     string
		expectedHex string
	}{
		{
			name:        "RFC4493 zero key empty message",
			keyHex:      "2b7e151628aed2a6abf7158809cf4f3c",
			dataHex:     "",
			expectedHex: "bb1d6929e95937287fa37d129b756746",
		},
		{
			name:        "RFC4493 16 byte message",
			keyHex:      "2b7e151628aed2a6abf7158809cf4f3c",
			dataHex:     "6bc1bee22e409f96e93d7e117393172a",
			expectedHex: "070a16b46b4d4144f79bdd9dd04a287c",
		},
		{
			name:        "RFC4493 40 byte message",
			keyHex:      "2b7e151628aed2a6abf7158809cf4f3c",
			dataHex:     "6bc1bee22e409f96e93d7e117393172aae2d8a571e03ac9c9eb76fac45af8e5130c81c46a35ce411",
			expectedHex: "dfa66747de9ae63030ca32611497c827",
		},
		{
			name:        "RFC4493 64 byte message",
			keyHex:      "2b7e151628aed2a6abf7158809cf4f3c",
			dataHex:     "6bc1bee22e409f96e93d7e117393172aae2d8a571e03ac9c9eb76fac45af8e5130c81c46a35ce411e5fbc1191a0a52eff69f2445df4f9b17ad2b417be66c3710",
			expectedHex: "51f0bebf7e3b9d92fc49741779363cfe",
		},
		{
			name:        "all zeros key and data",
			keyHex:      "00000000000000000000000000000000",
			dataHex:     "",
			expectedHex: "4387c14b46ef7e176dceefa862d72ff9",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key, err := hex.DecodeString(tc.keyHex)
			if err != nil {
				t.Fatalf("hex decode key: %v", err)
			}
			data, err := hex.DecodeString(tc.dataHex)
			if err != nil {
				t.Fatalf("hex decode data: %v", err)
			}

			result, err := Sum(key, data)
			if err != nil {
				t.Fatalf("Sum error: %v", err)
			}

			got := hex.EncodeToString(result)
			if got != tc.expectedHex {
				t.Fatalf("Sum(%s, %s) = %s, want %s", tc.keyHex, tc.dataHex, got, tc.expectedHex)
			}
		})
	}
}

func TestSumMethodVsPackageSum(t *testing.T) {
	cases := []struct {
		name string
		data []byte
	}{
		{
			name: "empty data",
			data: []byte{},
		},
		{
			name: "single byte",
			data: []byte{0x42},
		},
		{
			name: "partial block 7 bytes",
			data: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07},
		},
		{
			name: "exact block",
			data: make([]byte, 16),
		},
		{
			name: "two and a half blocks",
			data: make([]byte, 40),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key := make([]byte, 16)

			pkgResult, err := Sum(key, tc.data)
			if err != nil {
				t.Fatalf("Package Sum error: %v", err)
			}

			h, err := NewCMAC(key)
			if err != nil {
				t.Fatalf("NewCMAC error: %v", err)
			}
			if len(tc.data) > 0 {
				_, _ = h.Write(tc.data)
			}
			methodResult := h.Sum(nil)

			if !bytes.Equal(pkgResult, methodResult) {
				t.Fatalf("Package Sum (%x) != method Sum (%x)", pkgResult, methodResult)
			}
		})
	}
}

func TestWriteIncrementalVsSingleShot(t *testing.T) {
	cases := []struct {
		name   string
		chunks [][]byte
		full   []byte
	}{
		{
			name:   "two halves of a block",
			chunks: [][]byte{make([]byte, 8), make([]byte, 8)},
			full:   make([]byte, 16),
		},
		{
			name:   "byte by byte for short data",
			chunks: [][]byte{{0x01}, {0x02}, {0x03}, {0x04}},
			full:   []byte{0x01, 0x02, 0x03, 0x04},
		},
		{
			name:   "three chunks forming two blocks",
			chunks: [][]byte{make([]byte, 10), make([]byte, 10), make([]byte, 12)},
			full:   make([]byte, 32),
		},
		{
			name:   "single byte then rest of block",
			chunks: [][]byte{{0x00}, make([]byte, 15)},
			full:   make([]byte, 16),
		},
		{
			name: "many tiny writes",
			chunks: func() [][]byte {
				c := make([][]byte, 33)
				for i := range c {
					c[i] = []byte{0x00}
				}
				return c
			}(),
			full: make([]byte, 33),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key := make([]byte, 16)

			h1, _ := NewCMAC(key)
			for _, chunk := range tc.chunks {
				_, _ = h1.Write(chunk)
			}
			mac1 := h1.Sum(nil)

			h2, _ := NewCMAC(key)
			_, _ = h2.Write(tc.full)
			mac2 := h2.Sum(nil)

			if !bytes.Equal(mac1, mac2) {
				t.Fatalf("Incremental %x != single shot %x", mac1, mac2)
			}
		})
	}
}

func TestDifferentKeysProduceDifferentMacs(t *testing.T) {
	cases := []struct {
		name string
		data []byte
	}{
		{
			name: "empty data",
			data: []byte{},
		},
		{
			name: "single byte",
			data: []byte{0x42},
		},
		{
			name: "block aligned",
			data: make([]byte, 16),
		},
		{
			name: "multi block",
			data: make([]byte, 48),
		},
		{
			name: "partial block",
			data: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key1 := make([]byte, 16)
			key2 := make([]byte, 16)
			key2[0] = 0x01

			mac1, err := Sum(key1, tc.data)
			if err != nil {
				t.Fatalf("Sum with key1 error: %v", err)
			}
			mac2, err := Sum(key2, tc.data)
			if err != nil {
				t.Fatalf("Sum with key2 error: %v", err)
			}

			if bytes.Equal(mac1, mac2) {
				t.Fatalf("Different keys produced same MAC: %x", mac1)
			}
		})
	}
}

func TestDifferentDataProducesDifferentMacs(t *testing.T) {
	cases := []struct {
		name  string
		data1 []byte
		data2 []byte
	}{
		{
			name:  "empty vs single byte",
			data1: []byte{},
			data2: []byte{0x01},
		},
		{
			name:  "single byte vs two bytes",
			data1: []byte{0x01},
			data2: []byte{0x01, 0x02},
		},
		{
			name:  "one block vs two blocks",
			data1: make([]byte, 16),
			data2: make([]byte, 32),
		},
		{
			name:  "same length different content",
			data1: []byte{0x00, 0x00, 0x00, 0x00},
			data2: []byte{0x00, 0x00, 0x00, 0x01},
		},
		{
			name:  "15 bytes vs 16 bytes",
			data1: make([]byte, 15),
			data2: make([]byte, 16),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key := make([]byte, 16)
			mac1, _ := Sum(key, tc.data1)
			mac2, _ := Sum(key, tc.data2)

			if bytes.Equal(mac1, mac2) {
				t.Fatalf("Different data produced same MAC: %x", mac1)
			}
		})
	}
}

func TestSumOutputLength(t *testing.T) {
	cases := []struct {
		name    string
		keySize int
		dataLen int
	}{
		{
			name:    "16 byte key empty data",
			keySize: 16,
			dataLen: 0,
		},
		{
			name:    "16 byte key 1 byte data",
			keySize: 16,
			dataLen: 1,
		},
		{
			name:    "24 byte key 16 byte data",
			keySize: 24,
			dataLen: 16,
		},
		{
			name:    "32 byte key 100 byte data",
			keySize: 32,
			dataLen: 100,
		},
		{
			name:    "16 byte key 256 byte data",
			keySize: 16,
			dataLen: 256,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key := make([]byte, tc.keySize)
			data := make([]byte, tc.dataLen)
			result, err := Sum(key, data)
			if err != nil {
				t.Fatalf("Sum error: %v", err)
			}
			if len(result) != 16 {
				t.Fatalf("Sum output length = %d, want 16", len(result))
			}
		})
	}
}

func TestSumNilData(t *testing.T) {
	cases := []struct {
		name    string
		keySize int
	}{
		{
			name:    "16 byte key",
			keySize: 16,
		},
		{
			name:    "24 byte key",
			keySize: 24,
		},
		{
			name:    "32 byte key",
			keySize: 32,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key := make([]byte, tc.keySize)
			result, err := Sum(key, nil)
			if err != nil {
				t.Fatalf("Sum(nil) unexpected error: %v", err)
			}
			if len(result) != 16 {
				t.Fatalf("Sum(nil) output length = %d, want 16", len(result))
			}
		})
	}
}

func TestSumNilDataMatchesNoWrite(t *testing.T) {
	cases := []struct {
		name    string
		keySize int
	}{
		{
			name:    "16 byte key",
			keySize: 16,
		},
		{
			name:    "24 byte key",
			keySize: 24,
		},
		{
			name:    "32 byte key",
			keySize: 32,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key := make([]byte, tc.keySize)

			pkgResult, err := Sum(key, nil)
			if err != nil {
				t.Fatalf("Package Sum error: %v", err)
			}

			h, err := NewCMAC(key)
			if err != nil {
				t.Fatalf("NewCMAC error: %v", err)
			}
			methodResult := h.Sum(nil)

			if !bytes.Equal(pkgResult, methodResult) {
				t.Fatalf("Sum(nil) %x != no-write Sum %x", pkgResult, methodResult)
			}
		})
	}
}

func TestMultipleResets(t *testing.T) {
	cases := []struct {
		name      string
		numResets int
	}{
		{
			name:      "single reset",
			numResets: 1,
		},
		{
			name:      "double reset",
			numResets: 2,
		},
		{
			name:      "five resets",
			numResets: 5,
		},
		{
			name:      "ten resets",
			numResets: 10,
		},
		{
			name:      "reset after each sum",
			numResets: 3,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key := make([]byte, 16)
			h, err := NewCMAC(key)
			if err != nil {
				t.Fatalf("NewCMAC error: %v", err)
			}

			data := []byte{0x01, 0x02, 0x03}
			var firstMac []byte

			for i := 0; i < tc.numResets; i++ {
				_, _ = h.Write(data)
				mac := h.Sum(nil)
				if i == 0 {
					firstMac = mac
				} else {
					if !bytes.Equal(mac, firstMac) {
						t.Fatalf("Reset #%d produced different MAC: %x != %x", i, mac, firstMac)
					}
				}
				h.Reset()
			}
		})
	}
}
