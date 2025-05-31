package fuzzysearch

import (
	"testing"

	"github.com/InsideGallery/core/testutils"
)

func TestSearch(t *testing.T) {
	s := NewIndex()
	d := NewDocument(1, "test string")
	d2 := NewDocument(2, "some name")
	d3 := NewDocument(3, "awesome name")
	s.Add(d, d2, d3)
	testutils.Equal(t, s.Search("test"), []int{1})
}
