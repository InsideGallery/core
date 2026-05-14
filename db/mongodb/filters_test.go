package mongodb

import (
	"errors"
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestCoreOwnedFilters(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		operation func() (any, error)
		want      any
		wantErr   error
	}{
		{
			name: "new filter",
			operation: func() (any, error) {
				return NewFilter(Field{Name: "email", Value: "a@test"}), nil
			},
			want: Filter{"email": "a@test"},
		},
		{
			name: "filter from pairs",
			operation: func() (any, error) {
				return FilterFromPairs("email", "a@test")
			},
			want: Filter{"email": "a@test"},
		},
		{
			name: "filter from pairs rejects odd input",
			operation: func() (any, error) {
				return FilterFromPairs("email")
			},
			wantErr: ErrFilterPairCount,
		},
		{
			name: "filter from pairs rejects non strings key",
			operation: func() (any, error) {
				return FilterFromPairs(1, "a@test")
			},
			wantErr: ErrFilterKeyType,
		},
		{
			name: "new sort",
			operation: func() (any, error) {
				return NewSort(SortField{Name: "created_at", Descending: true}), nil
			},
			want: bson.D{{Key: "created_at", Value: sortDescending}},
		},
	}

	for _, test := range cases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := test.operation()
			if test.wantErr != nil {
				if !errors.Is(err, test.wantErr) {
					t.Fatalf("err = %v, want %v", err, test.wantErr)
				}

				return
			}

			if err != nil {
				t.Fatalf("operation() error: %v", err)
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("operation() = %#v, want %#v", got, test.want)
			}
		})
	}
}
