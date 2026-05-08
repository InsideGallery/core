package dataconv

import (
	"testing"
)

func TestMergeStructEdgeCases(t *testing.T) {
	type simple struct {
		A int
		B string
		C bool
	}

	cases := []struct {
		name  string
		dst   *simple
		src   simple
		wantA int
		wantB string
		wantC bool
	}{
		{"src_fills_empty_dst", &simple{}, simple{A: 1, B: "hello", C: true}, 1, "hello", true},
		{"dst_keeps_existing_values", &simple{A: 10, B: "world", C: true}, simple{A: 1, B: "hello", C: false}, 10, "world", true},
		{"partial_fill", &simple{A: 5}, simple{B: "test", C: true}, 5, "test", true},
		{"both_empty", &simple{}, simple{}, 0, "", false},
		{"src_zero_does_not_overwrite", &simple{A: 42, B: "keep"}, simple{A: 0, B: ""}, 42, "keep", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := MergeStruct(tc.dst, tc.src)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.dst.A != tc.wantA {
				t.Fatalf("A: expected %d, got %d", tc.wantA, tc.dst.A)
			}
			if tc.dst.B != tc.wantB {
				t.Fatalf("B: expected %q, got %q", tc.wantB, tc.dst.B)
			}
			if tc.dst.C != tc.wantC {
				t.Fatalf("C: expected %v, got %v", tc.wantC, tc.dst.C)
			}
		})
	}
}

func TestMergeStructWithSlices(t *testing.T) {
	type withSlice struct {
		Items []string
		Count int
	}

	cases := []struct {
		name      string
		dst       *withSlice
		src       withSlice
		wantItems []string
		wantCount int
	}{
		{"empty_dst_gets_src_slice", &withSlice{}, withSlice{Items: []string{"a", "b"}, Count: 2}, []string{"a", "b"}, 2},
		{"dst_keeps_existing_slice", &withSlice{Items: []string{"x"}}, withSlice{Items: []string{"y"}}, []string{"x"}, 0},
		{"nil_src_slice_no_overwrite", &withSlice{Items: []string{"a"}}, withSlice{}, []string{"a"}, 0},
		{"both_nil_slices", &withSlice{}, withSlice{}, nil, 0},
		{"dst_empty_slice_gets_src", &withSlice{Items: []string{}}, withSlice{Items: []string{"new"}}, []string{"new"}, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := MergeStruct(tc.dst, tc.src)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(tc.dst.Items) != len(tc.wantItems) {
				t.Fatalf("Items length: expected %d, got %d", len(tc.wantItems), len(tc.dst.Items))
			}
			for i := range tc.wantItems {
				if tc.dst.Items[i] != tc.wantItems[i] {
					t.Fatalf("Items[%d]: expected %q, got %q", i, tc.wantItems[i], tc.dst.Items[i])
				}
			}
			if tc.dst.Count != tc.wantCount {
				t.Fatalf("Count: expected %d, got %d", tc.wantCount, tc.dst.Count)
			}
		})
	}
}

func TestMergeStructWithMaps(t *testing.T) {
	cases := []struct {
		name     string
		dst      map[string]interface{}
		src      map[string]interface{}
		checkKey string
		wantVal  interface{}
	}{
		{"empty_dst_filled", map[string]interface{}{}, map[string]interface{}{"a": 1}, "a", 1},
		{"dst_keeps_existing", map[string]interface{}{"a": "old"}, map[string]interface{}{"a": "new"}, "a", "old"},
		{"new_key_added", map[string]interface{}{"a": 1}, map[string]interface{}{"b": 2}, "b", 2},
		{"both_empty", map[string]interface{}{}, map[string]interface{}{}, "a", nil},
		{"nil_value_in_src", map[string]interface{}{}, map[string]interface{}{"a": nil}, "a", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := MergeStruct(&tc.dst, tc.src)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got, exists := tc.dst[tc.checkKey]
			if tc.wantVal == nil {
				if exists && got != nil {
					t.Fatalf("expected nil or missing for key %q, got %v", tc.checkKey, got)
				}
				return
			}
			if got != tc.wantVal {
				t.Fatalf("key %q: expected %v, got %v", tc.checkKey, tc.wantVal, got)
			}
		})
	}
}

func TestMergeStructErrorCases(t *testing.T) {
	type simple struct {
		A int
	}

	cases := []struct {
		name    string
		dst     interface{}
		src     interface{}
		wantErr bool
	}{
		{"non_pointer_dst", simple{A: 1}, simple{A: 2}, true},
		{"pointer_dst_ok", &simple{A: 1}, simple{A: 2}, false},
		{"both_pointers_dst", &simple{}, simple{A: 5}, false},
		{"empty_structs", &simple{}, simple{}, false},
		{"src_fills_dst", &simple{}, simple{A: 10}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := MergeStruct(tc.dst, tc.src)
			if tc.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestMergeStructNestedStructs(t *testing.T) {
	type inner struct {
		X int
		Y string
	}
	type outer struct {
		Name  string
		Inner inner
	}

	cases := []struct {
		name  string
		dst   *outer
		src   outer
		wantX int
		wantY string
		wantN string
	}{
		{"nested_fill", &outer{}, outer{Name: "test", Inner: inner{X: 1, Y: "y"}}, 1, "y", "test"},
		{"nested_partial", &outer{Inner: inner{X: 5}}, outer{Inner: inner{Y: "hello"}}, 5, "hello", ""},
		{"nested_no_overwrite", &outer{Inner: inner{X: 10, Y: "keep"}}, outer{Inner: inner{X: 20, Y: "new"}}, 10, "keep", ""},
		{"outer_only", &outer{}, outer{Name: "abc"}, 0, "", "abc"},
		{"inner_only", &outer{}, outer{Inner: inner{X: 99}}, 99, "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := MergeStruct(tc.dst, tc.src)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.dst.Inner.X != tc.wantX {
				t.Fatalf("Inner.X: expected %d, got %d", tc.wantX, tc.dst.Inner.X)
			}
			if tc.dst.Inner.Y != tc.wantY {
				t.Fatalf("Inner.Y: expected %q, got %q", tc.wantY, tc.dst.Inner.Y)
			}
			if tc.dst.Name != tc.wantN {
				t.Fatalf("Name: expected %q, got %q", tc.wantN, tc.dst.Name)
			}
		})
	}
}
