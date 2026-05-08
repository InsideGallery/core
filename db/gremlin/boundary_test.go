package gremlin

import "testing"

func TestVertexStoreContract(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		assert func(t *testing.T)
	}{
		{
			name: "client implements vertex store",
			assert: func(t *testing.T) {
				t.Helper()

				var _ VertexStore = (*Client)(nil)
				var _ GraphStore = (*Client)(nil)
			},
		},
		{
			name: "properties are stable key value pairs",
			assert: func(t *testing.T) {
				t.Helper()

				got := propertiesKeyValues(map[string]any{"b": 2, "a": 1})
				want := []any{"a", 1, "b", 2}

				if len(got) != len(want) {
					t.Fatalf("len(propertiesKeyValues()) = %d, want %d", len(got), len(want))
				}

				for i := range want {
					if got[i] != want[i] {
						t.Fatalf("propertiesKeyValues()[%d] = %v, want %v", i, got[i], want[i])
					}
				}
			},
		},
	}

	for _, test := range cases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			test.assert(t)
		})
	}
}
