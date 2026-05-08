package neo4j

import "testing"

func TestGraphContract(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		assert func(t *testing.T)
	}{
		{
			name: "graph client implements graph",
			assert: func(t *testing.T) {
				t.Helper()

				var _ Graph = (*GraphClient)(nil)
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
