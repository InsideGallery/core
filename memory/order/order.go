// Package order provides ordering helpers for memory data structures.
//
// New code should import this package for sorting helpers:
//
//	import "github.com/InsideGallery/core/memory/order"
//
// Compatibility: memory/utils.Sort remains available for existing consumers.
// Prefer order.Sort in new code and keep additional ordering helpers near this
// package rather than extending the legacy aggregate path.
package order

import (
	"github.com/InsideGallery/core/memory/comparator"
	legacy "github.com/InsideGallery/core/memory/utils"
)

// Sort sorts values in place according to the comparator.
func Sort(values []interface{}, compare comparator.Comparator) {
	legacy.Sort(values, compare)
}
