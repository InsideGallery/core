package elasticsearch

import "testing"

func TestSearcherContract(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		assert func(t *testing.T)
	}{
		{
			name: "search client implements searcher",
			assert: func(t *testing.T) {
				t.Helper()

				var _ Searcher = (*SearchClient)(nil)
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
