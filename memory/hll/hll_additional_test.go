package hll

import "testing"

func TestHyperLogLogAdditionalPaths(t *testing.T) {
	first, err := New()
	if err != nil {
		t.Fatalf("new first hll: %v", err)
	}

	second, err := New()
	if err != nil {
		t.Fatalf("new second hll: %v", err)
	}

	first.Add([]byte("alpha"))
	second.Add([]byte("beta"))

	union, err := first.Union(second)
	if err != nil {
		t.Fatalf("union: %v", err)
	}

	if union.Count() != 2 {
		t.Fatalf("union count = %d, want 2", union.Count())
	}

	if _, err := FromBytes([]byte("bad")); err == nil {
		t.Fatal("expected from bytes error")
	}
}
