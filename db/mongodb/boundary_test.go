package mongodb

import (
	"errors"
	"testing"
)

func TestDocumentStoreContract(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		assert func(t *testing.T)
	}{
		{
			name: "mongo client implements document store",
			assert: func(t *testing.T) {
				t.Helper()

				var _ DocumentStore = (*MongoClient)(nil)
			},
		},
		{
			name: "missing target error is stable",
			assert: func(t *testing.T) {
				t.Helper()

				if !errors.Is(ErrDocumentTargetIsNotSet, ErrDocumentTargetIsNotSet) {
					t.Fatal("ErrDocumentTargetIsNotSet should match itself")
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
