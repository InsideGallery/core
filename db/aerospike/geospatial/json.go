package geospatial

import (
	"math"
	"strconv"
	"strings"
)

const (
	FloatPrecision = 8
	FloatBitSize   = 64
)

type GeoPoint struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

func (p *GeoPoint) GetPoint() Point {
	return NewPoint(p.Coordinates...)
}

// Vector describe any vector point
type Vector interface {
	Coordinate(i int) (r float64)
	Coordinates() [3]float64
}

// Point describe point
type Point struct {
	coordinates [3]float64
}

// NewPoint return new point
func NewPoint(s ...float64) Point {
	var x, y, z float64
	l := len(s)

	switch l {
	case 3: // nolint:mnd
		x, y, z = s[0], s[1], s[2]
	case 2: // nolint:mnd
		x, y = s[0], s[1]
	case 1: // nolint:mnd
		x = s[0]
	}

	return Point{
		coordinates: [3]float64{x, y, z},
	}
}

// Coordinate return coordinate for dimension (0-x, 1-y, 2-z)
func (p Point) Coordinate(i int) float64 {
	if i > 2 || i < 0 {
		return 0
	} // nolint:mnd

	return p.coordinates[i]
}

// Coordinates return all coordinates
func (p Point) Coordinates() [3]float64 {
	return p.coordinates
}

// Dot scalar multiply
func (p Point) Dot(p2 Vector) float64 {
	var sum float64

	for i, v := range p.Coordinates() {
		sum += v * p2.Coordinate(i)
	}

	return sum
}

// Normal returns the vector's norm.
func (p Point) Normal() float64 {
	return math.Sqrt(p.Dot(p))
}

// NormalSquare returns the vector's norm square.
func (p Point) NormalSquare() float64 {
	return p.Dot(p)
}

// DistanceSquare return distance square between positions (euclidean)
func (p Point) DistanceSquare(p2 Vector) float64 {
	var sum float64

	for i := range p.Coordinates() {
		d := p.Coordinate(i) - p2.Coordinate(i)
		sum += d * d
	}

	return sum
}

// Distance return distance between positions (euclidean)
func (p Point) Distance(p2 Vector) float64 {
	return math.Sqrt(p.DistanceSquare(p2))
}

func PointStr(lat, lon float64) string {
	return strings.Join([]string{
		"{\"type\": \"Point\", \"coordinates\": [",
		strconv.FormatFloat(lon, 'f', FloatPrecision, FloatBitSize),
		",",
		strconv.FormatFloat(lat, 'f', FloatPrecision, FloatBitSize),
		"]}",
	}, "")
}

func PolygonStr(points ...[]float64) string {
	rawPoints := make([]string, len(points))

	for i, p := range points {
		values := make([]string, len(p))
		for j, v := range p {
			values[j] = strconv.FormatFloat(v, 'f', FloatPrecision, FloatBitSize)
		}

		rawPoints[i] = strings.Join([]string{"[", strings.Join(values, ","), "]"}, "")
	}

	return strings.Join([]string{
		"{ \"type\": \"Polygon\", \"coordinates\": [[",
		strings.Join(rawPoints, ", "),
		"]] }",
	}, "")
}
