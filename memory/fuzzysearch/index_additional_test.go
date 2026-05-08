package fuzzysearch

import (
	"reflect"
	"testing"

	"github.com/InsideGallery/core/memory/set"
)

func TestIndexOperations(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "document terms are analyzed",
			run: func(t *testing.T) {
				t.Helper()

				doc := NewDocument(7, "Hello, WORLD!")
				if got := doc.Terms(); !reflect.DeepEqual(got, []string{"hello", "world"}) {
					t.Fatalf("terms = %#v", got)
				}
			},
		},
		{
			name: "add avoids duplicate document ids per token",
			run: func(t *testing.T) {
				t.Helper()

				index := NewIndex()
				index.Add(NewDocument(1, "go go"), NewDocument(2, "go fast"))

				if got := index["go"]; !reflect.DeepEqual(got, []int{1, 2}) {
					t.Fatalf("ids = %#v, want [1 2]", got)
				}
			},
		},
		{
			name: "remove deletes existing ids and ignores missing tokens",
			run: func(t *testing.T) {
				t.Helper()

				index := NewIndex()
				first := NewDocument(1, "go fast")
				second := NewDocument(2, "go slow")
				index.Add(first, second)
				index.Remove(first, NewDocument(3, "missing"))

				if got := index.Search("go"); !reflect.DeepEqual(got, []int{2}) {
					t.Fatalf("search = %#v, want [2]", got)
				}
			},
		},
		{
			name: "search intersects all tokens",
			run: func(t *testing.T) {
				t.Helper()

				index := NewIndex()
				index.Add(
					NewDocument(1, "red blue"),
					NewDocument(2, "red green"),
					NewDocument(3, "red blue green"),
				)

				if got := index.Search("red blue"); !reflect.DeepEqual(got, []int{1, 3}) {
					t.Fatalf("search = %#v, want [1 3]", got)
				}

				if got := index.Search("purple"); got != nil {
					t.Fatalf("missing search = %#v, want nil", got)
				}
			},
		},
		{
			name: "intersection handles uneven sorted inputs",
			run: func(t *testing.T) {
				t.Helper()

				got := Intersection([]int{1, 3, 5, 7}, []int{2, 3, 4, 7, 9})
				if !reflect.DeepEqual(got, []int{3, 7}) {
					t.Fatalf("intersection = %#v, want [3 7]", got)
				}
			},
		},
		{
			name: "stopword filter removes configured words",
			run: func(t *testing.T) {
				t.Helper()

				got := StopwordFilter([]string{"keep", "drop", "also"}, set.NewGenericDataSet("drop"))
				if !reflect.DeepEqual(got, []string{"keep", "also"}) {
					t.Fatalf("filtered = %#v, want [keep also]", got)
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}
