package comparator

import (
	"math"
	"testing"
	"time"
)

func TestStringComparator(t *testing.T) {
	cases := []struct {
		name string
		a    interface{}
		b    interface{}
		want int
	}{
		{"equal single char", "a", "a", 0},
		{"a less than b", "a", "b", -1},
		{"a greater than b", "b", "a", 1},
		{"both empty", "", "", 0},
		{"empty vs non-empty", "", "a", -1},
		{"non-empty vs empty", "a", "", 1},
		{"prefix shorter", "aa", "aab", -1},
		{"prefix longer", "aab", "aa", 1},
		{"equal long strings", "abcdefgh", "abcdefgh", 0},
		{"empty vs long string", "", "aaaaaaa", -1},
		{"long string vs empty", "aaaaaaa", "", 1},
		{"differ at first char", "z", "a", 1},
		{"differ at last char", "abc", "abd", -1},
		{"unicode bytes", "z", "A", 1},
		{"single char equal", "x", "x", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := StringComparator(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("StringComparator(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestIntComparator(t *testing.T) {
	cases := []struct {
		name string
		a    interface{}
		b    interface{}
		want int
	}{
		{"equal zero", 0, 0, 0},
		{"equal positive", 42, 42, 0},
		{"equal negative", -10, -10, 0},
		{"a less than b positive", 1, 2, -1},
		{"a greater than b positive", 2, 1, 1},
		{"negative less than zero", -1, 0, -1},
		{"zero greater than negative", 0, -1, 1},
		{"both negative a less", -5, -3, -1},
		{"both negative a greater", -3, -5, 1},
		{"large positive values", 1000000, 999999, 1},
		{"large negative values", -999999, -1000000, 1},
		{"min and max near", -1, 1, -1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := IntComparator(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("IntComparator(%v, %v) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestInt8Comparator(t *testing.T) {
	cases := []struct {
		name string
		a    interface{}
		b    interface{}
		want int
	}{
		{"equal zero", int8(0), int8(0), 0},
		{"equal positive", int8(10), int8(10), 0},
		{"equal negative", int8(-10), int8(-10), 0},
		{"a less than b", int8(1), int8(2), -1},
		{"a greater than b", int8(2), int8(1), 1},
		{"negative vs positive", int8(-1), int8(1), -1},
		{"positive vs negative", int8(1), int8(-1), 1},
		{"max value equal", int8(math.MaxInt8), int8(math.MaxInt8), 0},
		{"min value equal", int8(math.MinInt8), int8(math.MinInt8), 0},
		{"min vs max", int8(math.MinInt8), int8(math.MaxInt8), -1},
		{"max vs min", int8(math.MaxInt8), int8(math.MinInt8), 1},
		{"zero vs negative", int8(0), int8(-1), 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Int8Comparator(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("Int8Comparator(%v, %v) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestInt16Comparator(t *testing.T) {
	cases := []struct {
		name string
		a    interface{}
		b    interface{}
		want int
	}{
		{"equal zero", int16(0), int16(0), 0},
		{"equal positive", int16(100), int16(100), 0},
		{"equal negative", int16(-100), int16(-100), 0},
		{"a less than b", int16(1), int16(2), -1},
		{"a greater than b", int16(2), int16(1), 1},
		{"negative vs positive", int16(-1), int16(1), -1},
		{"max value equal", int16(math.MaxInt16), int16(math.MaxInt16), 0},
		{"min value equal", int16(math.MinInt16), int16(math.MinInt16), 0},
		{"min vs max", int16(math.MinInt16), int16(math.MaxInt16), -1},
		{"max vs min", int16(math.MaxInt16), int16(math.MinInt16), 1},
		{"zero vs max", int16(0), int16(math.MaxInt16), -1},
		{"zero vs min", int16(0), int16(math.MinInt16), 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Int16Comparator(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("Int16Comparator(%v, %v) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestInt32Comparator(t *testing.T) {
	cases := []struct {
		name string
		a    interface{}
		b    interface{}
		want int
	}{
		{"equal zero", int32(0), int32(0), 0},
		{"equal positive", int32(500), int32(500), 0},
		{"equal negative", int32(-500), int32(-500), 0},
		{"a less than b", int32(1), int32(2), -1},
		{"a greater than b", int32(2), int32(1), 1},
		{"negative vs positive", int32(-50), int32(50), -1},
		{"max value equal", int32(math.MaxInt32), int32(math.MaxInt32), 0},
		{"min value equal", int32(math.MinInt32), int32(math.MinInt32), 0},
		{"min vs max", int32(math.MinInt32), int32(math.MaxInt32), -1},
		{"max vs min", int32(math.MaxInt32), int32(math.MinInt32), 1},
		{"zero vs negative", int32(0), int32(-1), 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Int32Comparator(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("Int32Comparator(%v, %v) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestInt64Comparator(t *testing.T) {
	cases := []struct {
		name string
		a    interface{}
		b    interface{}
		want int
	}{
		{"equal zero", int64(0), int64(0), 0},
		{"equal positive", int64(999), int64(999), 0},
		{"equal negative", int64(-999), int64(-999), 0},
		{"a less than b", int64(1), int64(2), -1},
		{"a greater than b", int64(2), int64(1), 1},
		{"negative vs positive", int64(-100), int64(100), -1},
		{"max value equal", int64(math.MaxInt64), int64(math.MaxInt64), 0},
		{"min value equal", int64(math.MinInt64), int64(math.MinInt64), 0},
		{"min vs max", int64(math.MinInt64), int64(math.MaxInt64), -1},
		{"max vs min", int64(math.MaxInt64), int64(math.MinInt64), 1},
		{"zero vs max", int64(0), int64(math.MaxInt64), -1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Int64Comparator(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("Int64Comparator(%v, %v) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestUIntComparator(t *testing.T) {
	cases := []struct {
		name string
		a    interface{}
		b    interface{}
		want int
	}{
		{"equal zero", uint(0), uint(0), 0},
		{"equal positive", uint(42), uint(42), 0},
		{"a less than b", uint(1), uint(2), -1},
		{"a greater than b", uint(2), uint(1), 1},
		{"zero vs one", uint(0), uint(1), -1},
		{"one vs zero", uint(1), uint(0), 1},
		{"large values equal", uint(math.MaxUint32), uint(math.MaxUint32), 0},
		{"large vs small", uint(1000000), uint(1), 1},
		{"small vs large", uint(1), uint(1000000), -1},
		{"adjacent values", uint(99), uint(100), -1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := UIntComparator(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("UIntComparator(%v, %v) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestUInt8Comparator(t *testing.T) {
	cases := []struct {
		name string
		a    interface{}
		b    interface{}
		want int
	}{
		{"equal zero", uint8(0), uint8(0), 0},
		{"equal positive", uint8(100), uint8(100), 0},
		{"a less than b", uint8(1), uint8(2), -1},
		{"a greater than b", uint8(2), uint8(1), 1},
		{"zero vs max", uint8(0), uint8(math.MaxUint8), -1},
		{"max vs zero", uint8(math.MaxUint8), uint8(0), 1},
		{"max equal", uint8(math.MaxUint8), uint8(math.MaxUint8), 0},
		{"adjacent values", uint8(127), uint8(128), -1},
		{"zero vs one", uint8(0), uint8(1), -1},
		{"one vs zero", uint8(1), uint8(0), 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := UInt8Comparator(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("UInt8Comparator(%v, %v) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestUInt16Comparator(t *testing.T) {
	cases := []struct {
		name string
		a    interface{}
		b    interface{}
		want int
	}{
		{"equal zero", uint16(0), uint16(0), 0},
		{"equal positive", uint16(300), uint16(300), 0},
		{"a less than b", uint16(1), uint16(2), -1},
		{"a greater than b", uint16(2), uint16(1), 1},
		{"zero vs max", uint16(0), uint16(math.MaxUint16), -1},
		{"max vs zero", uint16(math.MaxUint16), uint16(0), 1},
		{"max equal", uint16(math.MaxUint16), uint16(math.MaxUint16), 0},
		{"zero vs one", uint16(0), uint16(1), -1},
		{"large vs small", uint16(60000), uint16(100), 1},
		{"adjacent at boundary", uint16(255), uint16(256), -1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := UInt16Comparator(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("UInt16Comparator(%v, %v) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestUInt32Comparator(t *testing.T) {
	cases := []struct {
		name string
		a    interface{}
		b    interface{}
		want int
	}{
		{"equal zero", uint32(0), uint32(0), 0},
		{"equal positive", uint32(12345), uint32(12345), 0},
		{"a less than b", uint32(1), uint32(2), -1},
		{"a greater than b", uint32(2), uint32(1), 1},
		{"zero vs max", uint32(0), uint32(math.MaxUint32), -1},
		{"max vs zero", uint32(math.MaxUint32), uint32(0), 1},
		{"max equal", uint32(math.MaxUint32), uint32(math.MaxUint32), 0},
		{"zero vs one", uint32(0), uint32(1), -1},
		{"large values", uint32(math.MaxUint32 - 1), uint32(math.MaxUint32), -1},
		{"one vs zero", uint32(1), uint32(0), 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := UInt32Comparator(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("UInt32Comparator(%v, %v) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestUInt64Comparator(t *testing.T) {
	cases := []struct {
		name string
		a    interface{}
		b    interface{}
		want int
	}{
		{"equal zero", uint64(0), uint64(0), 0},
		{"equal positive", uint64(99999), uint64(99999), 0},
		{"a less than b", uint64(1), uint64(2), -1},
		{"a greater than b", uint64(2), uint64(1), 1},
		{"zero vs max", uint64(0), uint64(math.MaxUint64), -1},
		{"max vs zero", uint64(math.MaxUint64), uint64(0), 1},
		{"max equal", uint64(math.MaxUint64), uint64(math.MaxUint64), 0},
		{"zero vs one", uint64(0), uint64(1), -1},
		{"adjacent at max", uint64(math.MaxUint64 - 1), uint64(math.MaxUint64), -1},
		{"one vs zero", uint64(1), uint64(0), 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := UInt64Comparator(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("UInt64Comparator(%v, %v) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestFloat32Comparator(t *testing.T) {
	cases := []struct {
		name string
		a    interface{}
		b    interface{}
		want int
	}{
		{"equal zero", float32(0), float32(0), 0},
		{"equal positive", float32(1.5), float32(1.5), 0},
		{"equal negative", float32(-1.5), float32(-1.5), 0},
		{"a less than b", float32(1.0), float32(2.0), -1},
		{"a greater than b", float32(2.0), float32(1.0), 1},
		{"negative vs positive", float32(-1.0), float32(1.0), -1},
		{"positive vs negative", float32(1.0), float32(-1.0), 1},
		{"zero vs negative", float32(0), float32(-0.5), 1},
		{"small positive difference", float32(1.0001), float32(1.0), 1},
		{"max equal", float32(math.MaxFloat32), float32(math.MaxFloat32), 0},
		{"smallest positive vs zero", float32(math.SmallestNonzeroFloat32), float32(0), 1},
		{"negative zero vs positive zero", math.Float32frombits(1 << 31), float32(0.0), 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Float32Comparator(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("Float32Comparator(%v, %v) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestFloat64Comparator(t *testing.T) {
	cases := []struct {
		name string
		a    interface{}
		b    interface{}
		want int
	}{
		{"equal zero", float64(0), float64(0), 0},
		{"equal positive", float64(3.14), float64(3.14), 0},
		{"equal negative", float64(-3.14), float64(-3.14), 0},
		{"a less than b", float64(1.0), float64(2.0), -1},
		{"a greater than b", float64(2.0), float64(1.0), 1},
		{"negative vs positive", float64(-1.0), float64(1.0), -1},
		{"positive vs negative", float64(1.0), float64(-1.0), 1},
		{"zero vs negative", float64(0), float64(-0.5), 1},
		{"max equal", math.MaxFloat64, math.MaxFloat64, 0},
		{"smallest positive vs zero", math.SmallestNonzeroFloat64, float64(0), 1},
		{"negative zero vs positive zero", math.Copysign(0, -1), float64(0), 0},
		{"very close values", float64(1.0000000001), float64(1.0000000002), -1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Float64Comparator(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("Float64Comparator(%v, %v) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestFloat64DiffComparator(t *testing.T) {
	cases := []struct {
		name string
		a    interface{}
		b    interface{}
		want int
	}{
		{"equal zero", float64(0), float64(0), 0},
		{"equal positive", float64(5.0), float64(5.0), 0},
		{"equal negative", float64(-5.0), float64(-5.0), 0},
		{"a greater than b large diff", float64(10.0), float64(1.0), 1},
		{"a less than b negative diff treated as zero", float64(1.0), float64(10.0), 0},
		{"diff at smallest nonzero boundary", float64(math.SmallestNonzeroFloat64), float64(0), 0},
		{"negative zero vs positive zero", math.Copysign(0, -1), float64(0), 0},
		{"very small positive diff", float64(1.0) + math.SmallestNonzeroFloat64, float64(1.0), 0},
		{"negative diff treated as zero", float64(-1e100), float64(1e100), 0},
		{"large positive vs large negative", float64(1e100), float64(-1e100), 1},
		{"slightly above smallest nonzero", float64(0) + 2*math.SmallestNonzeroFloat64, float64(0), 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Float64DiffComparator(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("Float64DiffComparator(%v, %v) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestByteComparator(t *testing.T) {
	cases := []struct {
		name string
		a    interface{}
		b    interface{}
		want int
	}{
		{"equal zero", byte(0), byte(0), 0},
		{"equal positive", byte(100), byte(100), 0},
		{"a less than b", byte(1), byte(2), -1},
		{"a greater than b", byte(2), byte(1), 1},
		{"zero vs max", byte(0), byte(math.MaxUint8), -1},
		{"max vs zero", byte(math.MaxUint8), byte(0), 1},
		{"max equal", byte(math.MaxUint8), byte(math.MaxUint8), 0},
		{"adjacent values", byte(127), byte(128), -1},
		{"zero vs one", byte(0), byte(1), -1},
		{"ascii A vs B", byte('A'), byte('B'), -1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ByteComparator(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("ByteComparator(%v, %v) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestRuneComparator(t *testing.T) {
	cases := []struct {
		name string
		a    interface{}
		b    interface{}
		want int
	}{
		{"equal zero", rune(0), rune(0), 0},
		{"equal ascii", rune('a'), rune('a'), 0},
		{"a less than b ascii", rune('a'), rune('b'), -1},
		{"a greater than b ascii", rune('b'), rune('a'), 1},
		{"negative vs positive", rune(-1), rune(1), -1},
		{"positive vs negative", rune(1), rune(-1), 1},
		{"zero vs positive", rune(0), rune(100), -1},
		{"unicode runes equal", rune(0x1F600), rune(0x1F600), 0},
		{"unicode rune less", rune(0x1F600), rune(0x1F601), -1},
		{"max int32 equal", rune(math.MaxInt32), rune(math.MaxInt32), 0},
		{"min int32 vs max int32", rune(math.MinInt32), rune(math.MaxInt32), -1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := RuneComparator(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("RuneComparator(%v, %v) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestTimeComparator(t *testing.T) {
	now := time.Now()
	epoch := time.Unix(0, 0)
	zeroTime := time.Time{}

	cases := []struct {
		name string
		a    interface{}
		b    interface{}
		want int
	}{
		{"equal now", now, now, 0},
		{"a after b", now.Add(24 * time.Hour), now, 1},
		{"a before b", now, now.Add(24 * time.Hour), -1},
		{"equal epoch", epoch, epoch, 0},
		{"epoch vs now", epoch, now, -1},
		{"now vs epoch", now, epoch, 1},
		{"zero time equal", zeroTime, zeroTime, 0},
		{"zero vs epoch", zeroTime, epoch, -1},
		{"nanosecond difference a after", now.Add(1 * time.Nanosecond), now, 1},
		{"nanosecond difference a before", now, now.Add(1 * time.Nanosecond), -1},
		{"two weeks apart", now.Add(14 * 24 * time.Hour), now, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := TimeComparator(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("TimeComparator(%v, %v) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}
