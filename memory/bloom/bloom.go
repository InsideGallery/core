package bloom

import (
	"encoding/binary"
	"fmt"
	"hash"
	"hash/fnv"
	"math"

	"github.com/InsideGallery/core/dataconv"
)

type filter struct {
	m uint32
	k uint32
	h hash.Hash64
}

func (f *filter) bits(data []byte) []uint32 {
	f.h.Reset()
	f.h.Write(data)
	d := f.h.Sum(nil)
	a := binary.BigEndian.Uint32(d[4:8])
	b := binary.BigEndian.Uint32(d[0:4])
	is := make([]uint32, f.k)

	for i := uint32(0); i < f.k; i++ {
		is[i] = (a + b*i) % f.m
	}

	return is
}

func newFilter(m, k uint32) *filter {
	return &filter{
		m: m,
		k: k,
		h: fnv.New64(),
	}
}

func estimates(n uint32, p float64) (uint32, uint32) {
	nf := float64(n)
	log2 := math.Log(2) // nolint:mnd
	m := -1 * nf * math.Log(p) / (log2 * log2)
	k := math.Ceil(log2 * m / nf)

	words := m + 31>>5 // nolint:mnd
	if words >= math.MaxInt32 {
		panic(fmt.Sprintf("A 32-bit bloom filter with n %d and p %f requires a "+
			"32-bit bitset with a slice of %f words, but slices cannot contain more than "+
			"%d elements. Please use the equivalent 64-bit bloom filter, e.g. New64(), "+
			"instead.", n, p, words, math.MaxInt32-1))
	} else if m > math.MaxUint32 {
		panic(fmt.Sprintf("A 32-bit bloom filter with n %d and p %f requires a "+
			"32-bit bitset with %f bits, but this number overflows an uint32. Please use "+
			"the equivalent 64-bit bloom filter, e.g. New64(), instead.", n, p, m))
	}

	return uint32(m), uint32(k)
}

type CountingFilter struct {
	*filter
	counters []byte
}

// NewCounting creates an optimized counting bloom filter.
func NewCounting(n int, p float64) *CountingFilter {
	m, k := estimates(uint32(n), p)

	return &CountingFilter{
		filter:   newFilter(m, k),
		counters: make([]byte, m),
	}
}

// Test checks if an item is likely in the set.
// It returns true if the item might be in the filter, false if it is definitely not.
func (f *CountingFilter) Test(data []byte) bool {
	for _, i := range f.bits(data) {
		if f.counters[i] == 0 {
			return false
		}
	}

	return true
}

// Add inserts data into the filter by incrementing its counters.
// It protects against counter overflow (stopping at 255).
func (f *CountingFilter) Add(data []byte) {
	for _, i := range f.bits(data) {
		// Prevent overflow
		if f.counters[i] < 255 { // nolint:mnd
			f.counters[i]++
		}
	}
}

// Remove deletes data from the filter by decrementing its counters.
// It protects against counter underflow (stopping at 0).
func (f *CountingFilter) Remove(data []byte) {
	for _, i := range f.bits(data) {
		// Prevent underflow
		if f.counters[i] > 0 {
			f.counters[i]--
		}
	}
}

// Reset clears all counters in the filter.
func (f *CountingFilter) Reset() {
	// More efficient than re-allocating memory
	for i := range f.counters {
		f.counters[i] = 0
	}
}

// ToBytes serializes the CountingFilter to a byte slice.
func (f *CountingFilter) ToBytes() ([]byte, error) {
	enc := dataconv.NewBinaryEncoder()

	err := enc.Encode(f.m)
	if err != nil {
		return nil, fmt.Errorf("encode m param: %w", err)
	}

	err = enc.Encode(f.k)
	if err != nil {
		return nil, fmt.Errorf("encode k param: %w", err)
	}

	err = enc.Encode(f.counters)
	if err != nil {
		return nil, fmt.Errorf("encode counters param: %w", err)
	}

	return enc.Bytes(), nil
}

// NewCountingFromBytes deserializes a byte slice into a CountingFilter.
func NewCountingFromBytes(data []byte) (*CountingFilter, error) {
	if len(data) == 0 {
		return nil, ErrEmptyDump
	}

	dec := dataconv.NewBinaryDecoder(data)

	res := &CountingFilter{
		filter: &filter{
			h: fnv.New64(),
		},
	}

	err := dec.Decode(&res.m)
	if err != nil {
		return nil, fmt.Errorf("decode m param: %w", err)
	}

	err = dec.Decode(&res.k)
	if err != nil {
		return nil, fmt.Errorf("decode k param: %w", err)
	}

	err = dec.Decode(&res.counters)
	if err != nil {
		return nil, fmt.Errorf("decode counters param: %w", err)
	}

	return res, nil
}
