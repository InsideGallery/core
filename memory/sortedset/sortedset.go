package sortedset

import (
	"math/rand"
	"sync"

	"github.com/InsideGallery/core/memory/comparator"
)

const (
	skipListMaxLevel = 32
	skipListP        = 0.25
	maxLimit         = 2147483648
)

type SortedSet[K comparable, V comparable] struct {
	emptyKey   K
	header     *Node[K, V]
	tail       *Node[K, V]
	dict       map[V]*Node[K, V]
	comparator comparator.Comparator
	length     uint64
	level      int
	mu         sync.RWMutex
}

func (s *SortedSet[K, V]) createNode(level int, key K, value V) *Node[K, V] {
	node := &Node[K, V]{
		key:   key,
		value: value,
		level: make([]Level[K, V], level),
	}

	return node
}

// RandomLevel returns a random level for the new skiplist node we are going to create.
// The return value of this function is between 1 and skipListMaxLevel
// (both inclusive), with a powerlaw-alike distribution where higher
// levels are less likely to be returned.
func (s *SortedSet[K, V]) randomLevel() int {
	level := 1
	for float64(rand.Int31()&0xFFFF) < float64(skipListP*0xFFFF) { //nolint:gosec,mnd
		level++
	}

	if level < skipListMaxLevel {
		return level
	}

	return skipListMaxLevel
}

func (s *SortedSet[K, V]) insertNode(key K, value V) *Node[K, V] {
	var update [skipListMaxLevel]*Node[K, V]
	var rank [skipListMaxLevel]uint64

	x := s.header

	for i := s.level - 1; i >= 0; i-- {
		/* store rank that is crossed to reach the insert position */
		if s.level-1 == i {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}

		for x.level[i].forward != nil &&
			(s.comparator(x.level[i].forward.key, key) < 0 ||
				(s.comparator(x.level[i].forward.key, key) == 0 && // key is the same but the key is different
					x.level[i].forward.value != value)) {
			rank[i] += x.level[i].span
			x = x.level[i].forward
		}
		update[i] = x
	}

	/* we assume the key is not already inside, since we allow duplicated
	 * keys, and the re-insertion of key and redis object should never
	 * happen since the caller of Insert() should test in the hash table
	 * if the element is already inside or not. */
	level := s.randomLevel()

	if level > s.level { // add a new level
		for i := s.level; i < level; i++ {
			rank[i] = 0
			update[i] = s.header
			update[i].level[i].span = s.length
		}
		s.level = level
	}

	x = s.createNode(level, key, value)
	for i := 0; i < level; i++ {
		x.level[i].forward = update[i].level[i].forward
		update[i].level[i].forward = x

		/* update span covered by update[i] as x is inserted here */
		x.level[i].span = update[i].level[i].span - (rank[0] - rank[i])
		update[i].level[i].span = (rank[0] - rank[i]) + 1
	}

	/* increment span for untouched levels */
	for i := level; i < s.level; i++ {
		update[i].level[i].span++
	}

	if update[0] == s.header {
		x.backward = nil
	} else {
		x.backward = update[0]
	}

	if x.level[0].forward != nil {
		x.level[0].forward.backward = x
	} else {
		s.tail = x
	}

	s.length++

	return x
}

func (s *SortedSet[K, V]) deleteNode(x *Node[K, V], update [skipListMaxLevel]*Node[K, V]) {
	for i := 0; i < s.level; i++ {
		if update[i].level[i].forward == x {
			update[i].level[i].span += x.level[i].span - 1
			update[i].level[i].forward = x.level[i].forward
		} else {
			update[i].level[i].span--
		}
	}

	if x.level[0].forward != nil {
		x.level[0].forward.backward = x.backward
	} else {
		s.tail = x.backward
	}

	for s.level > 1 && s.header.level[s.level-1].forward == nil {
		s.level--
	}

	s.length--
	delete(s.dict, x.value)
}

func (s *SortedSet[K, V]) delete(key K, value V) bool {
	var update [skipListMaxLevel]*Node[K, V]

	x := s.header
	for i := s.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && s.comparator(x.level[i].forward.key, key) < 0 {
			x = x.level[i].forward
		}
		update[i] = x
	}
	/* We may have multiple elements with the same key, what we need
	 * is to find the element with both the right key and object. */
	x = x.level[0].forward
	if x != nil && key == x.key && x.value == value {
		s.deleteNode(x, update)
		// free x
		return true
	}

	return false /* not found */
}

func NewSortedSet[K comparable, V comparable](c comparator.Comparator) *SortedSet[K, V] {
	var emptyKey K
	var emptyValue V

	sortedSet := &SortedSet[K, V]{
		level:      1,
		dict:       make(map[V]*Node[K, V]),
		comparator: c,
		emptyKey:   emptyKey,
	}
	sortedSet.header = sortedSet.createNode(skipListMaxLevel, emptyKey, emptyValue)

	return sortedSet
}

func (s *SortedSet[K, V]) GetCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	l := int(s.length)

	return l
}

// PeekMin get the element with minimum key, nil if the set is empty
// Time complexity of this method is : O(log(N))
func (s *SortedSet[K, V]) PeekMin() *Node[K, V] {
	s.mu.RLock()
	defer s.mu.RUnlock()

	f := s.header.level[0].forward

	return f
}

// PopMin get and remove the element with minimal key, nil if the set is empty
// Time complexity of this method is : O(log(N))
func (s *SortedSet[K, V]) PopMin() *Node[K, V] {
	s.mu.Lock()
	defer s.mu.Unlock()

	x := s.header.level[0].forward
	if x != nil {
		s.Remove(x.value)
	}

	return x
}

// PeekMax get the element with maximum key, nil if the set is empty
// Time Complexity : O(1)
func (s *SortedSet[K, V]) PeekMax() *Node[K, V] {
	s.mu.RLock()
	defer s.mu.RUnlock()

	t := s.tail

	return t
}

// PopMax get and remove the element with maximum key, nil if the set is empty
// Time complexity of this method is : O(log(N))
func (s *SortedSet[K, V]) PopMax() *Node[K, V] {
	s.mu.Lock()
	defer s.mu.Unlock()

	x := s.tail
	if x != nil {
		s.Remove(x.value)
	}

	return x
}

// Upsert add an element into the sorted set with specific key / value / key.
// if the element is added, this method returns true; otherwise false means updated
// Time complexity of this method is : O(log(N))
func (s *SortedSet[K, V]) Upsert(key K, value V) bool {
	var newNode *Node[K, V]
	s.mu.Lock()
	defer s.mu.Unlock()

	found := s.dict[value]
	if found != nil {
		// key does not change, only update value
		if s.comparator(found.key, key) == 0 {
			found.value = value
		} else { // key changes, delete and re-insert
			s.delete(found.key, found.value)
			newNode = s.insertNode(key, value)
		}
	} else {
		newNode = s.insertNode(key, value)
	}

	if newNode != nil {
		s.dict[value] = newNode
	}

	return found == nil
}

// Remove delete element specified by key
// Time complexity of this method is : O(log(N))
func (s *SortedSet[K, V]) Remove(value V) *Node[K, V] {
	s.mu.Lock()
	defer s.mu.Unlock()

	found := s.dict[value]
	if found != nil {
		s.delete(found.key, found.value)
		return found
	}

	return nil
}

// GetTop return top data
func (s *SortedSet[K, V]) GetTop(count int, remove bool) (result []*Node[K, V]) {
	return s.GetByRankRange(-1, -count, remove)
}

// GetRTop return top from end data
func (s *SortedSet[K, V]) GetRTop(count int, remove bool) (result []*Node[K, V]) {
	return s.GetByRankRange(1, count, remove)
}

// GetUntilKey get all values until given key
func (s *SortedSet[K, V]) GetUntilKey(untilKey K, remove bool) []any {
	nodes := s.GetByKeyRange(s.emptyKey, untilKey, &GetByKeyRangeOptions{
		Remove: remove,
	})

	data := make([]any, len(nodes))
	for i, nd := range nodes {
		data[i] = nd.Value()
	}

	return data
}

// GetByKeyRange get the nodes whose key within the specific range
// If options is nil, it `searches` in interval [start, end] without any limit by default
// Time complexity of this method is : O(log(N))
func (s *SortedSet[K, V]) GetByKeyRange(start K, end K, options *GetByKeyRangeOptions) []*Node[K, V] { //nolint:gocyclo
	s.mu.Lock()
	defer s.mu.Unlock()

	// prepare parameters
	limit := maxLimit
	if options != nil && options.Limit > 0 {
		limit = options.Limit
	}

	var remove bool
	if options != nil {
		remove = options.Remove
	}

	excludeStart := options != nil && options.ExcludeStart
	excludeEnd := options != nil && options.ExcludeEnd

	reverse := s.comparator(start, end) > 0
	if reverse {
		start, end = end, start
		excludeStart, excludeEnd = excludeEnd, excludeStart
	}

	var nodes []*Node[K, V]

	// determine if out of range
	if s.length == 0 {
		return nodes
	}

	if reverse { // search from end to start
		x := s.header

		if excludeEnd {
			for i := s.level - 1; i >= 0; i-- {
				for x.level[i].forward != nil &&
					s.comparator(x.level[i].forward.key, end) < 0 {
					x = x.level[i].forward
				}
			}
		} else {
			for i := s.level - 1; i >= 0; i-- {
				for x.level[i].forward != nil &&
					s.comparator(x.level[i].forward.key, end) <= 0 {
					x = x.level[i].forward
				}
			}
		}

		for x != nil && limit > 0 {
			if excludeStart {
				if s.comparator(x.key, start) <= 0 {
					break
				}
			} else {
				if s.comparator(x.key, start) < 0 {
					break
				}
			}

			next := x.backward
			nodes = append(nodes, x)

			if remove {
				s.delete(x.Key(), x.Value())
			}

			limit--
			x = next
		}
	} else {
		// search from start to end
		x := s.header

		if excludeStart {
			for i := s.level - 1; i >= 0; i-- {
				for x.level[i].forward != nil &&
					s.comparator(x.level[i].forward.key, start) <= 0 {
					x = x.level[i].forward
				}
			}
		} else {
			for i := s.level - 1; i >= 0; i-- {
				for x.level[i].forward != nil &&
					s.comparator(x.level[i].forward.key, start) < 0 {
					x = x.level[i].forward
				}
			}
		}

		/* Current node is the last with key < or <= start. */
		x = x.level[0].forward

		for x != nil && limit > 0 {
			if excludeEnd {
				if s.comparator(x.key, end) >= 0 {
					break
				}
			} else {
				if s.comparator(x.key, end) > 0 {
					break
				}
			}

			next := x.level[0].forward
			nodes = append(nodes, x)

			if remove {
				s.delete(x.Key(), x.Value())
			}

			limit--
			x = next
		}
	}

	return nodes
}

// GetByRankRange get nodes within specific rank range [start, end]
// Note that the rank is 1-based integer. Rank 1 means the first node; Rank -1 means the last node;
// If start is greater than end, the returned array is in reserved order
// If remove is true, the returned nodes are removed
// Time complexity of this method is : O(log(N))
func (s *SortedSet[K, V]) GetByRankRange(start int, end int, remove bool) []*Node[K, V] {
	s.mu.Lock()
	defer s.mu.Unlock()

	/* Sanitize indexes. */
	if start < 0 {
		start = int(s.length) + start + 1
	}

	if end < 0 {
		end = int(s.length) + end + 1
	}

	if start <= 0 {
		start = 1
	}

	if end <= 0 {
		end = 1
	}

	reverse := start > end
	if reverse { // swap start and end
		start, end = end, start
	}

	var update [skipListMaxLevel]*Node[K, V]
	var nodes []*Node[K, V]
	traversed := 0

	x := s.header

	for i := s.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil &&
			traversed+int(x.level[i].span) < start {
			traversed += int(x.level[i].span)
			x = x.level[i].forward
		}

		if remove {
			update[i] = x
		} else if traversed+1 == start {
			break
		}
	}

	traversed++
	x = x.level[0].forward

	for x != nil && traversed <= end {
		next := x.level[0].forward
		nodes = append(nodes, x)

		if remove {
			s.deleteNode(x, update)
		}

		traversed++
		x = next
	}

	if reverse {
		for i, j := 0, len(nodes)-1; i < j; i, j = i+1, j-1 {
			nodes[i], nodes[j] = nodes[j], nodes[i]
		}
	}

	return nodes
}

// GetByRank get  node by rank.
// Note that the rank is 1-based integer. Rank 1 means the first node; Rank -1 means the last node;
// If remove is true, the returned nodes are removed
// If node is not found at specific rank, nil is returned
// Time complexity of this method is : O(log(N))
func (s *SortedSet[K, V]) GetByRank(rank int, remove bool) *Node[K, V] {
	nodes := s.GetByRankRange(rank, rank, remove)
	if len(nodes) == 1 {
		return nodes[0]
	}

	return nil
}

// GetByValue get node by value
// If node is not found, nil is returned
// Time complexity : O(1)
func (s *SortedSet[K, V]) GetByValue(value V) *Node[K, V] {
	s.mu.RLock()
	defer s.mu.RUnlock()

	n := s.dict[value]

	return n
}

func (s *SortedSet[K, V]) Contains(value V) bool {
	return s.GetByValue(value) != nil
}

// FindRank find the rank of the node specified by key
// Note that the rank is 1-based integer. Rank 1 means the first node
// If the node is not found, 0 is returned. Otherwise rank(> 0) is returned
// Time complexity of this method is : O(log(N))
func (s *SortedSet[K, V]) FindRank(value V) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rank := 0
	node := s.dict[value]

	if node != nil {
		x := s.header
		for i := s.level - 1; i >= 0; i-- {
			for x.level[i].forward != nil &&
				(s.comparator(x.level[i].forward.key, node.key) < 0 ||
					(s.comparator(x.level[i].forward.key, node.key) == 0 &&
						x.level[i].forward.value != node.value)) {
				rank += int(x.level[i].span)
				x = x.level[i].forward
			}

			if x.value == value {
				return rank
			}
		}
	}

	return 0
}
