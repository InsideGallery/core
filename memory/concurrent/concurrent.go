// Package concurrent provides in-memory containers guarded for concurrent access.
//
// New code should import this package for safe list and map containers:
//
//	import "github.com/InsideGallery/core/memory/concurrent"
//
// Compatibility: the legacy memory/utils package still exposes SafeList and
// SafeMap. Prefer NewList and NewMap here so new consumers do not depend on the
// older aggregate helper path.
package concurrent

import legacy "github.com/InsideGallery/core/memory/utils"

// List is a concurrency-safe slice-like container.
type List[V any] = legacy.SafeList[V]

// Map is a concurrency-safe string-keyed map container.
type Map[K string, V any] = legacy.SafeMap[K, V]

// NewList returns a concurrency-safe list seeded with data.
func NewList[V any](data ...V) *List[V] {
	return legacy.NewSafeList(data...)
}

// NewMap returns a concurrency-safe map seeded with data.
func NewMap[K string, V any](data map[K]V) *Map[K, V] {
	return legacy.NewSafeMap(data)
}
