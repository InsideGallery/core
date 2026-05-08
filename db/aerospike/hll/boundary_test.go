package hll

import (
	"context"
	"errors"
	"testing"

	as "github.com/aerospike/aerospike-client-go/v7"
)

func TestCounterBoundary(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		counter *OperatorCounter
		options CountOptions
		want    CountResult
		wantErr error
	}{
		{
			name:    "empty keys return zero",
			counter: NewCounter(fakeOperator{}),
			options: CountOptions{
				Namespace: "test",
				Set:       "hll",
				Mode:      CountModeUnion,
			},
			want: CountResult{},
		},
		{
			name:    "missing operator returns stable error",
			counter: NewCounter(nil),
			wantErr: ErrCounterNotSet,
		},
	}

	for _, test := range cases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var _ Counter = (*OperatorCounter)(nil)

			got, err := test.counter.Count(context.Background(), test.options)
			if test.wantErr != nil {
				if !errors.Is(err, test.wantErr) {
					t.Fatalf("err = %v, want %v", err, test.wantErr)
				}

				return
			}

			if err != nil {
				t.Fatalf("Count() error: %v", err)
			}

			if got != test.want {
				t.Fatalf("Count() = %v, want %v", got, test.want)
			}
		})
	}
}

type fakeOperator struct{}

func (fakeOperator) Operate(*as.WritePolicy, *as.Key, ...*as.Operation) (*as.Record, as.Error) {
	panic("fake operator should not be called")
}
