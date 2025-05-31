//go:build unit
// +build unit

package geospatial

import (
	"testing"

	"github.com/InsideGallery/core/testutils"
)

func TestPoint(t *testing.T) {
	testcases := map[string]struct {
		lat float64
		lon float64
		r   string
	}{
		"point 12.20000000,15.12345679": {
			lon: 12.2,
			lat: 15.123456789,
			r:   "{\"type\": \"Point\", \"coordinates\": [12.20000000,15.12345679]}",
		},
	}

	for name, test := range testcases {
		test := test
		t.Run(name, func(t *testing.T) {
			testutils.Equal(t, PointStr(test.lat, test.lon), test.r)
		})
	}
}

func TestPolygon(t *testing.T) {
	testcases := map[string]struct {
		p1 []float64
		p2 []float64
		p3 []float64
		p4 []float64
		p5 []float64
		r  string
	}{
		"point many points": {
			p1: []float64{0, 0},
			p2: []float64{0, 10},
			p3: []float64{10, 10},
			p4: []float64{10, 0},
			p5: []float64{0, 0},
			r:  "{ \"type\": \"Polygon\", \"coordinates\": [[[0.00000000,0.00000000], [0.00000000,10.00000000], [10.00000000,10.00000000], [10.00000000,0.00000000], [0.00000000,0.00000000]]] }",
		},
	}

	for name, test := range testcases {
		test := test
		t.Run(name, func(t *testing.T) {
			testutils.Equal(t, PolygonStr(test.p1, test.p2, test.p3, test.p4, test.p5), test.r)
		})
	}
}
