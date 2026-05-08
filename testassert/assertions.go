// Package testassert provides shared assertion helpers for tests.
//
// New code should import this package instead of the legacy testutils path:
//
//	import "github.com/InsideGallery/core/testassert"
//
//	testassert.Equal(t, got, want)
//
// Compatibility: github.com/InsideGallery/core/testutils remains available for
// existing tests. Prefer testassert in new tests so assertions have a focused
// package name and do not extend the legacy utility aggregate.
package testassert

import (
	"testing"

	"github.com/InsideGallery/core/testutils"
)

// ApproximatelyEqual reports whether two floats are nearly equal.
func ApproximatelyEqual(a, b float64) bool {
	return testutils.ApproximatelyEqual(a, b)
}

// EqualError reports whether two errors are equivalent.
func EqualError(received, expected error) bool {
	return testutils.EqualError(received, expected)
}

// IsEqual reports whether two values are equivalent.
func IsEqual(received, expected interface{}) bool {
	return testutils.IsEqual(received, expected)
}

// NotEqual fails the test when two values are equivalent.
func NotEqual(t testing.TB, received, expected interface{}) {
	t.Helper()

	testutils.NotEqual(t, received, expected)
}

// Equal fails the test when two values are not equivalent.
func Equal(t testing.TB, received, expected interface{}) {
	t.Helper()

	testutils.Equal(t, received, expected)
}

// EqualJSON fails the test when two JSON values are not equivalent.
func EqualJSON(t testing.TB, received, expected []byte) {
	t.Helper()

	testutils.EqualJSON(t, received, expected)
}

// NotEqualJSON fails the test when two JSON values are equivalent.
func NotEqualJSON(t testing.TB, received, expected []byte) {
	t.Helper()

	testutils.NotEqualJSON(t, received, expected)
}
