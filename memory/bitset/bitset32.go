// nolint:mnd
package bitset

import (
	"bytes"
	"fmt"
	"math"

	"github.com/InsideGallery/core/dataconv"
)

const (
	sw32   uint32 = 32
	slg232 uint32 = 5
	m132   uint32 = 0x55555555 // 0101...
	m232   uint32 = 0x33333333 // 00110011...
	m432   uint32 = 0x0f0f0f0f // 00001111...
	hff32  uint32 = 0xffffffff // all ones
)

func wordsNeeded32(n uint32) uint32 {
	if n == 0 {
		return 1
	} else if n > math.MaxUint32-sw32 {
		return math.MaxUint32 >> slg232
	}

	return (n + (sw32 - 1)) >> slg232
}

type Bitset32 struct {
	n uint32
	b []uint32
}

// Returns the current size of the bitset.
func (b *Bitset32) Len() uint32 {
	return b.n
}

// Test whether bit i is set.
func (b *Bitset32) Test(i uint32) bool {
	if i >= b.n {
		return false
	}

	return ((b.b[i>>slg232] & (1 << (i & (sw32 - 1)))) != 0)
}

// Set bit i to 1.
func (b *Bitset32) Set(i uint32) {
	if i >= b.n {
		nsize := wordsNeeded32(i + 1)
		l := uint32(len(b.b))

		if nsize > l {
			nb := make([]uint32, nsize-l)
			b.b = append(b.b, nb...)
		}

		b.n = i + 1
	}

	b.b[i>>slg232] |= (1 << (i & (sw32 - 1)))
}

// Set bit i to 0.
func (b *Bitset32) Clear(i uint32) {
	if i >= b.n {
		return
	}

	b.b[i>>slg232] &^= 1 << (i & (sw32 - 1))
}

// Flip bit i.
func (b *Bitset32) Flip(i uint32) {
	if i >= b.n {
		b.Set(i)
	}

	b.b[i>>slg232] ^= 1 << (i & (sw32 - 1))
}

// Clear all bits in the bitset.
func (b *Bitset32) Reset() {
	for i := range b.b {
		b.b[i] = 0
	}
}

// Get the number of words used in the bitset.
func (b *Bitset32) wordCount() uint32 {
	return wordsNeeded32(b.n)
}

// Clone the bitset.
func (b *Bitset32) Clone() *Bitset32 {
	c := New32(b.n)
	copy(c.b, b.b)

	return c
}

// Copy the bitset into another bitset, returning the size of the destination
// bitset.
func (b *Bitset32) Copy(c *Bitset32) (n uint32) {
	copy(c.b, b.b)

	n = c.n
	if b.n < c.n {
		n = b.n
	}

	return
}

func popCountUint32(x uint32) uint32 {
	x -= (x >> 1) & m132               // put count of each 2 bits into those 2 bits
	x = (x & m232) + ((x >> 2) & m232) // put count of each 4 bits into those 4 bits
	x = (x + (x >> 4)) & m432          // put count of each 8 bits into those 8 bits
	x += x >> 8                        // put count of each 16 bits into their lowest 8 bits
	x += x >> 16                       // put count of each 32 bits into their lowest 8 bits

	return x & 0x7f
}

// Get the number of set bits in the bitset.
func (b *Bitset32) Count() uint32 {
	sum := uint32(0)
	for _, w := range b.b {
		sum += popCountUint32(w)
	}

	return sum
}

// Test if two bitsets are equal. Returns true if both bitsets are the same
// size and all the same bits are set in both bitsets.
func (b *Bitset32) Equal(c *Bitset32) bool {
	if b.n != c.n {
		return false
	}

	for p, v := range b.b {
		if c.b[p] != v {
			return false
		}
	}

	return true
}

// Bitset &^ (and or); difference between receiver and another set.
func (b *Bitset32) Difference(ob *Bitset32) (result *Bitset32) {
	result = b.Clone() // clone b (in case b is bigger than ob)
	szl := ob.wordCount()
	l := uint32(len(b.b))

	for i := uint32(0); i < l; i++ {
		if i >= szl {
			break
		}

		result.b[i] = b.b[i] &^ ob.b[i]
	}

	return
}

func sortByLength32(a *Bitset32, b *Bitset32) (ap *Bitset32, bp *Bitset32) {
	if a.n <= b.n {
		ap, bp = a, b
	} else {
		ap, bp = b, a
	}

	return
}

// Bitset & (and); intersection of receiver and another set.
func (b *Bitset32) Intersection(ob *Bitset32) (result *Bitset32) {
	b, ob = sortByLength32(b, ob)
	result = New32(b.n)

	for i, w := range b.b {
		result.b[i] = w & ob.b[i]
	}

	return
}

// Bitset | (or); union of receiver and another set.
func (b *Bitset32) Union(ob *Bitset32) (result *Bitset32) {
	b, ob = sortByLength32(b, ob)
	result = ob.Clone()
	szl := ob.wordCount()
	l := uint32(len(b.b))

	for i := uint32(0); i < l; i++ {
		if i >= szl {
			break
		}

		result.b[i] = b.b[i] | ob.b[i]
	}

	return
}

// Bitset ^ (xor); symmetric difference of receiver and another set.
func (b *Bitset32) SymmetricDifference(ob *Bitset32) (result *Bitset32) {
	b, ob = sortByLength32(b, ob)
	// ob is bigger, so clone it
	result = ob.Clone()
	szl := b.wordCount()
	l := uint32(len(b.b))

	for i := uint32(0); i < l; i++ {
		if i >= szl {
			break
		}

		result.b[i] = b.b[i] ^ ob.b[i]
	}

	return
}

// Return true if the bitset's length is a multiple of the word size.
func (b *Bitset32) isEven() bool {
	return (b.n % sw32) == 0
}

// Clean last word by setting unused bits to 0.
func (b *Bitset32) cleanLastWord() {
	if !b.isEven() {
		b.b[wordsNeeded32(b.n)-1] &= (hff32 >> (sw32 - (b.n % sw32)))
	}
}

// Return the (local) complement of a bitset (up to n bits).
func (b *Bitset32) Complement() (result *Bitset32) {
	result = New32(b.n)
	for i, w := range b.b {
		result.b[i] = ^(w)
	}

	result.cleanLastWord()

	return
}

// Returns true if all bits in the bitset are set.
func (b *Bitset32) All() bool {
	return b.Count() == b.n
}

// Returns true if no bit in the bitset is set.
func (b *Bitset32) None() bool {
	for _, w := range b.b {
		if w > 0 {
			return false
		}
	}

	return true
}

// Return true if any bit in the bitset is set.
func (b *Bitset32) Any() bool {
	return !b.None()
}

// Get a string representation of the words in the bitset.
func (b *Bitset32) String() string {
	buffer := bytes.NewBufferString("")
	for i := int(wordsNeeded32(b.n) - 1); i >= 0; i-- {
		fmt.Fprintf(buffer, "%032b.", b.b[i])
	}

	return buffer.String()
}

// Make a new bitset with a starting capacity of n bits. The bitset expands
// automatically.
func New32(n uint32) *Bitset32 {
	nWords := wordsNeeded32(n)
	if nWords > math.MaxInt32-1 {
		panic(fmt.Sprintf("Bitset32 needs %d %d-bit words to store %d bits, but slices cannot "+
			"hold more than %d items. Please use a Bitset64 instead.", nWords, sw32, n, math.MaxInt32-1))
	}

	b := &Bitset32{
		n,
		make([]uint32, nWords),
	}

	return b
}

// ToBytes serializes the Bitset32 to a byte slice.
func (b *Bitset32) ToBytes() ([]byte, error) {
	enc := dataconv.NewBinaryEncoder()

	err := enc.Encode(b.n)
	if err != nil {
		return nil, fmt.Errorf("encode n param: %w", err)
	}

	err = enc.Encode(len(b.b))
	if err != nil {
		return nil, fmt.Errorf("encode size param: %w", err)
	}

	for _, v := range b.b {
		err = enc.Encode(v)
		if err != nil {
			return nil, fmt.Errorf("encode b param: %w, v: %v", err, v)
		}
	}

	raw := enc.Bytes()

	return raw, nil
}

// NewFromBytes32 deserializes a byte slice into a Bitset32.
func NewFromBytes32(data []byte) (*Bitset32, error) {
	dec := dataconv.NewBinaryDecoder(data)

	res := &Bitset32{}

	var size int

	err := dec.Decode(&res.n)
	if err != nil {
		return nil, fmt.Errorf("decode n param: %w", err)
	}

	err = dec.Decode(&size)
	if err != nil {
		return nil, fmt.Errorf("decode size param: %w", err)
	}

	for i := 0; i < size; i++ {
		var val uint32

		err = dec.Decode(&val)
		if err != nil {
			return nil, fmt.Errorf("decode b param: %w", err)
		}

		res.b = append(res.b, val)
	}

	return res, nil
}
