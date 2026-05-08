package aerospike

import "testing"

func TestNamespaceStoreContract(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		assert func(t *testing.T)
	}{
		{
			name: "namespace instance implements namespace store",
			assert: func(t *testing.T) {
				t.Helper()

				var _ NamespaceStore = (*NamespaceInstance)(nil)
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
