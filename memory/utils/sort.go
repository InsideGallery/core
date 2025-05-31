package utils

import (
	"sort"

	"github.com/InsideGallery/core/memory/comparator"
)

// Sort sorts values (in-place) with respect to the given comparator.
//
// Uses Go's sort (hybrid of quicksort for large and then insertion sort for smaller slices).
func Sort(values []interface{}, comparator comparator.Comparator) {
	sort.Sort(sortable{values: values, comparator: comparator})
}

type sortable struct {
	comparator comparator.Comparator
	values     []interface{}
}

// Len return len of slice
func (s sortable) Len() int {
	return len(s.values)
}

// Swap implement swap items
func (s sortable) Swap(i, j int) {
	s.values[i], s.values[j] = s.values[j], s.values[i]
}

// Less return true if i-th less of j-th item
func (s sortable) Less(i, j int) bool {
	return s.comparator(s.values[i], s.values[j]) < 0
}
