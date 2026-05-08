package subscriber

import "testing"

func TestGenericSubscriberBoundaryAliases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		assert func(t *testing.T)
	}{
		{
			name: "aliases compile",
			assert: func(t *testing.T) {
				t.Helper()

				var _ Client
				var _ Message
				var _ MessageHandler
				var _ SubscriptionHandle
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
