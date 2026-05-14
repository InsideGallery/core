# Aerospike Geospatial

Import path: `github.com/InsideGallery/core/db/aerospike/geospatial`

This package provides small helpers for Aerospike-compatible GeoJSON strings
and three-dimensional point math.

## Main APIs

- `GeoPoint` models a GeoJSON point with `type` and `coordinates` fields.
- `GeoPoint.GetPoint` converts a `GeoPoint` to `Point`.
- `Vector` is the coordinate-reading interface used by point math.
- `Point` stores up to three coordinates.
- `NewPoint`, `Coordinate`, and `Coordinates` create and inspect points.
- `Dot`, `Normal`, `NormalSquare`, `DistanceSquare`, and `Distance` perform vector math.
- `PointStr(lat, lon)` formats a GeoJSON point string.
- `PolygonStr(points...)` formats a GeoJSON polygon string from coordinate slices.
- `FloatPrecision` and `FloatBitSize` control float formatting; current output uses 8 decimal places.

## Usage

```go
package example

import "github.com/InsideGallery/core/db/aerospike/geospatial"

func geoValues() (string, string, float64) {
	point := geospatial.PointStr(15.123456789, 12.2)
	polygon := geospatial.PolygonStr(
		[]float64{0, 0},
		[]float64{0, 10},
		[]float64{10, 10},
		[]float64{10, 0},
		[]float64{0, 0},
	)
	distance := geospatial.NewPoint(0, 0).Distance(geospatial.NewPoint(3, 4))

	return point, polygon, distance
}
```

`PointStr` takes latitude first and longitude second, then writes GeoJSON
coordinates as `[lon,lat]`. `PolygonStr` writes points exactly as supplied; the
caller is responsible for closing the polygon ring when needed.

## Operational Notes

The package does not open Aerospike connections and does not validate GeoJSON
geometry. `NewPoint` fills missing dimensions with zero, ignores dimensions
after the third, and `Coordinate` returns zero for indexes outside `0..2`.
