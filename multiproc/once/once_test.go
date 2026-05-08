package once

import (
	"errors"
	"testing"
)

func TestOnce(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("try again")

	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "success runs once until reset",
			run: func(t *testing.T) {
				t.Helper()

				var (
					once  Once
					calls int
				)

				for i := 0; i < 2; i++ {
					err := once.Do(func() error {
						calls++

						return nil
					})
					if err != nil {
						t.Fatalf("do: %v", err)
					}
				}

				if calls != 1 {
					t.Fatalf("calls = %d, want 1", calls)
				}

				once.Reset()

				err := once.Do(func() error {
					calls++

					return nil
				})
				if err != nil {
					t.Fatalf("do after reset: %v", err)
				}

				if calls != 2 {
					t.Fatalf("calls after reset = %d, want 2", calls)
				}
			},
		},
		{
			name: "error keeps once retryable",
			run: func(t *testing.T) {
				t.Helper()

				var once Once

				err := once.Do(func() error {
					return expectedErr
				})
				if !errors.Is(err, expectedErr) {
					t.Fatalf("err = %v, want %v", err, expectedErr)
				}

				err = once.Do(func() error {
					return nil
				})
				if err != nil {
					t.Fatalf("retry do: %v", err)
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}
