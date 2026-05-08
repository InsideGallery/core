package redis

import "testing"

func TestKeyValueStoreContract(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		assert func(t *testing.T)
	}{
		{
			name: "connection implements key value store",
			assert: func(t *testing.T) {
				t.Helper()

				var _ KeyValueStore = Connection{}
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
