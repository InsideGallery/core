//go:build unit
// +build unit

package mathutils

import (
	"testing"

	"github.com/InsideGallery/core/testutils"
)

func TestWeightIndex(t *testing.T) {
	result := map[interface{}]int{}

	for i := 0; i < 10000; i++ {
		r := WeightIndex(map[interface{}]uint64{
			"p1":  100,
			"p2":  100,
			"p3":  100,
			"p4":  100,
			"p5":  100,
			"p6":  100,
			"p7":  100,
			"p8":  100,
			"p9":  100,
			"p10": 100,
		})

		if _, e := result[r]; !e {
			result[r] = 0
		}
		result[r]++
	}
	v := float64(result["p1"]) / float64(result["p2"])

	testutils.Equal(t, v > 0.7, true)
	testutils.Equal(t, v <= 1.3, true)
}

/*
goos: linux
goarch: amd64
BenchmarkWeightIndex-12                  2202054               539 ns/op              56 B/op          4 allocs/op
PASS
*/
var res interface{}

func BenchmarkWeightIndex(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			result := WeightIndex(map[interface{}]uint64{
				"p1": 50,
				"p2": 50,
			})
			res = result
		}
	})
}
