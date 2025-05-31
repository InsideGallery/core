package testutils

import (
	"encoding/json"
	nerrors "errors"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/pkg/errors"
)

// ApproximatelyEqual function to test if two real numbers are (almost) equal
func ApproximatelyEqual(a, b float64) bool {
	epsilon := math.SmallestNonzeroFloat64
	difference := a - b

	return difference < epsilon && difference > -epsilon
}

// EqualError return true if two errors are equal
func EqualError(v, e error) bool {
	return nerrors.Is(v, e) || errors.Cause(v) == errors.Cause(e) || strings.EqualFold(v.Error(), e.Error())
}

// IsEqual check if two interface equal
func IsEqual(received, expected interface{}) bool {
	switch v := received.(type) {
	case float64:
		e, ok := expected.(float64)
		if !ok {
			return false
		}

		return ApproximatelyEqual(v, e)
	case float32:
		e, ok := expected.(float32)
		if !ok {
			return false
		}

		return ApproximatelyEqual(float64(v), float64(e))
	case error:
		e, ok := expected.(error)
		if !ok {
			return false
		}

		return EqualError(v, e)
	default:
		return reflect.DeepEqual(received, expected)
	}
}

// NotEqual checks if two variables are not equal
func NotEqual(t testing.TB, received, expected interface{}) {
	t.Helper()

	if IsEqual(received, expected) {
		t.Fatalf("mismatched values, received and expected should not be equal: %v", received)
	}
}

// Equal checks if two variables are equal
func Equal(t testing.TB, received, expected interface{}) {
	t.Helper()

	if !IsEqual(received, expected) {
		t.Fatalf("mismatched values: %v != %v", received, expected)
	}
}

// EqualJSON compare JSONs
func EqualJSON(t testing.TB, received, expected []byte) {
	t.Helper()

	if len(expected) == 0 {
		if len(received) != 0 {
			t.Fatalf("expected empty string, got: %s", received)
		}

		return
	}

	rec := map[string]interface{}{}
	if err := json.Unmarshal(received, &rec); err != nil {
		t.Fatalf("Unable to parse json: %s", received)
	}

	exp := map[string]interface{}{}
	if err := json.Unmarshal(expected, &exp); err != nil {
		t.Fatalf("Unable to parse json: %s", expected)
	}

	Equal(t, rec, exp)
}

// NotEqualJSON compare JSONs
func NotEqualJSON(t testing.TB, received, expected []byte) {
	t.Helper()

	if len(expected) == 0 {
		if len(received) == 0 {
			t.Fatal("expected and received are empty")
		}

		return
	}

	rec := map[string]interface{}{}
	if err := json.Unmarshal(received, &rec); err != nil {
		t.Fatalf("Unable to parse json: %s", received)
	}

	exp := map[string]interface{}{}
	if err := json.Unmarshal(expected, &exp); err != nil {
		t.Fatalf("Unable to parse json: %s", expected)
	}

	NotEqual(t, rec, exp)
}
