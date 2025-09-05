package bloom

import (
	"fmt"
	"hash"
	"hash/crc64"
	"hash/fnv"
	"math"

	"github.com/InsideGallery/core/dataconv"
	"github.com/InsideGallery/core/memory/bitset"
)

type filter64 struct {
	m  uint64
	k  uint64
	h  hash.Hash64
	oh hash.Hash64
}

func (f *filter64) bits(data []byte) []uint64 {
	f.h.Reset()
	f.h.Write(data)
	a := f.h.Sum64()

	f.oh.Reset()
	f.oh.Write(data)
	b := f.oh.Sum64()

	is := make([]uint64, f.k)
	for i := uint64(0); i < f.k; i++ {
		is[i] = (a + b*i) % f.m
	}

	return is
}

func newFilter64(m, k uint64) *filter64 {
	return &filter64{
		m:  m,
		k:  k,
		h:  fnv.New64(),
		oh: crc64.New(crc64.MakeTable(crc64.ECMA)),
	}
}

func estimates64(n uint64, p float64) (uint64, uint64) {
	nf := float64(n)
	log2 := math.Log(2) // nolint:mnd
	m := -1 * nf * math.Log(p) / log2 * log2
	k := math.Ceil(log2 * m / nf)

	return uint64(m), uint64(k)
}

// A standard 64-bit bloom filter using the 64-bit FNV-1a hash function.
type Filter64 struct {
	*filter64
	b *bitset.Bitset64
}

// Check whether data was previously added to the filter. Returns true if
// yes, with a false positive chance near the ratio specified upon creation
// of the filter. The result cannot be falsely negative.
func (f *Filter64) Test(data []byte) bool {
	for _, i := range f.bits(data) {
		if !f.b.Test(i) {
			return false
		}
	}

	return true
}

// Add data to the filter.
func (f *Filter64) Add(data []byte) {
	for _, i := range f.bits(data) {
		f.b.Set(i)
	}
}

// Resets the filter.
func (f *Filter64) Reset() {
	f.b.Reset()
}

// Create a bloom filter with an expected n number of items, and an acceptable
// false positive rate of p, e.g. 0.01 for 1%.
func New64(n int64, p float64) *Filter64 {
	m, k := estimates64(uint64(n), p)
	f := &Filter64{
		newFilter64(m, k),
		bitset.New64(m),
	}

	return f
}

// A counting bloom filter using the 64-bit FNV-1a hash function. Supports
// removing items from the filter.
type CountingFilter64 struct {
	*filter64
	b []*bitset.Bitset64
}

// Checks whether data was previously added to the filter. Returns true if
// yes, with a false positive chance near the ratio specified upon creation
// of the filter. The result cannot cannot be falsely negative (unless one
// has removed an item that wasn't actually added to the filter previously.)
func (f *CountingFilter64) Test(data []byte) bool {
	b := f.b[0]
	for _, v := range f.bits(data) {
		if !b.Test(v) {
			return false
		}
	}

	return true
}

// Adds data to the filter.
func (f *CountingFilter64) Add(data []byte) {
	for _, v := range f.bits(data) {
		done := false

		for _, ov := range f.b {
			if !ov.Test(v) {
				done = true

				ov.Set(v)

				break
			}
		}

		if !done {
			nb := bitset.New64(f.b[0].Len())
			f.b = append(f.b, nb)
			nb.Set(v)
		}
	}
}

// Removes data from the filter. This exact data must have been previously added
// to the filter, or future results will be inconsistent.
func (f *CountingFilter64) Remove(data []byte) {
	last := len(f.b) - 1

	for _, v := range f.bits(data) {
		for oi := last; oi >= 0; oi-- {
			ov := f.b[oi]
			if ov.Test(v) {
				ov.Clear(v)
				break
			}
		}
	}
}

// Resets the filter.
func (f *CountingFilter64) Reset() {
	f.b = f.b[:1]
	f.b[0].Reset()
}

// Create a counting bloom filter with an expected n number of items, and an
// acceptable false positive rate of p. Counting bloom filters support
// the removal of items from the filter.
func NewCounting64(n int64, p float64) *CountingFilter64 {
	m, k := estimates64(uint64(n), p)
	f := &CountingFilter64{
		newFilter64(m, k),
		[]*bitset.Bitset64{bitset.New64(m)},
	}

	return f
}

// A layered bloom filter using the 64-bit FNV-1a hash function.
type LayeredFilter64 struct {
	*filter64
	b []*bitset.Bitset64
}

// Checks whether data was previously added to the filter. Returns the number of
// the last layer where the data was added, e.g. 1 for the first layer, and a
// boolean indicating whether the data was added to the filter at all. The check
// has a false positive chance near the ratio specified upon creation of the
// filter. The result cannot be falsely negative.
func (f *LayeredFilter64) Test(data []byte) (int, bool) {
	is := f.bits(data)

	for i := len(f.b) - 1; i >= 0; i-- {
		v := f.b[i]
		last := len(is) - 1

		for oi, ov := range is {
			if !v.Test(ov) {
				break
			}

			if oi == last {
				// Every test was positive at this layer
				return i + 1, true
			}
		}
	}

	return 0, false
}

// Adds data to the filter. Returns the number of the layer where the data
// was added, e.g. 1 for the first layer.
func (f *LayeredFilter64) Add(data []byte) int {
	is := f.bits(data)

	var (
		i int
		v *bitset.Bitset64
	)

	for i, v = range f.b {
		here := false
		for _, ov := range is {
			if here {
				v.Set(ov)
			} else if !v.Test(ov) {
				here = true

				v.Set(ov)
			}
		}

		if here {
			return i + 1
		}
	}

	nb := bitset.New64(f.b[0].Len())
	f.b = append(f.b, nb)

	for _, v := range is {
		nb.Set(v)
	}

	return i + 2 // nolint:mnd
}

// Resets the filter.
func (f *LayeredFilter64) Reset() {
	f.b = f.b[:1]
	f.b[0].Reset()
}

// Create a layered bloom filter with an expected n number of items, and an
// acceptable false positive rate of p. Layered bloom filters can be used
// to keep track of a certain, arbitrary count of items, e.g. to check if some
// given data was added to the filter 10 times or less.
func NewLayered64(n int64, p float64) *LayeredFilter64 {
	m, k := estimates64(uint64(n), p)
	f := &LayeredFilter64{
		newFilter64(m, k),
		[]*bitset.Bitset64{bitset.New64(m)},
	}

	return f
}

// ToBytes serializes the CountingFilter to a byte slice.
func (f *CountingFilter64) ToBytes() ([]byte, error) {
	enc := dataconv.NewBinaryEncoder()

	err := enc.Encode(f.m)
	if err != nil {
		return nil, fmt.Errorf("encode m param: %w", err)
	}

	err = enc.Encode(f.k)
	if err != nil {
		return nil, fmt.Errorf("encode k param: %w", err)
	}

	err = enc.Encode(len(f.b))
	if err != nil {
		return nil, fmt.Errorf("encode size param: %w", err)
	}

	for _, v := range f.b {
		res, err := v.ToBytes()
		if err != nil {
			return nil, fmt.Errorf("get raw bitset param: %w", err)
		}

		err = enc.Encode(res)
		if err != nil {
			return nil, fmt.Errorf("encode raw bitset param: %w", err)
		}
	}

	return enc.Bytes(), nil
}

// NewCounting64FromBytes deserializes a byte slice into a CountingFilter.
func NewCounting64FromBytes(data []byte) (*CountingFilter64, error) {
	if len(data) == 0 {
		return nil, ErrEmptyDump
	}

	dec := dataconv.NewBinaryDecoder(data)

	res := &CountingFilter64{
		filter64: &filter64{
			h:  fnv.New64(),
			oh: crc64.New(crc64.MakeTable(crc64.ECMA)),
		},
	}

	var size int

	err := dec.Decode(&res.m)
	if err != nil {
		return nil, fmt.Errorf("decode m param: %w", err)
	}

	err = dec.Decode(&res.k)
	if err != nil {
		return nil, fmt.Errorf("decode k param: %w", err)
	}

	err = dec.Decode(&size)
	if err != nil {
		return nil, fmt.Errorf("decode size param: %w", err)
	}

	for i := 0; i < size; i++ {
		var raw []byte

		err = dec.Decode(&raw)
		if err != nil {
			return nil, fmt.Errorf("decode raw bitset param: %w", err)
		}

		b64, err := bitset.NewFromBytes64(raw)
		if err != nil {
			return nil, fmt.Errorf("decode create new bitset param: %w", err)
		}

		res.b = append(res.b, b64)
	}

	return res, nil
}
