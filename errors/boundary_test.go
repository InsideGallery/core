package errors //nolint:revive

import (
	nativeErrors "errors"
	"testing"
)

func TestWrapBoundary(t *testing.T) {
	t.Parallel()

	errBase := nativeErrors.New("sdk failed")

	cases := []struct {
		name     string
		err      error
		wantNil  bool
		wantText string
	}{
		{
			name:    "nil error returns nil",
			err:     nil,
			wantNil: true,
		},
		{
			name:     "non nil error is wrapped",
			err:      errBase,
			wantText: "mongodb: find: sdk failed",
		},
	}

	for _, test := range cases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := WrapBoundary("mongodb", "find", test.err)
			if test.wantNil {
				if got != nil {
					t.Fatalf("WrapBoundary() = %v, want nil", got)
				}

				return
			}

			if got.Error() != test.wantText {
				t.Fatalf("WrapBoundary().Error() = %q, want %q", got.Error(), test.wantText)
			}

			if !nativeErrors.Is(got, errBase) {
				t.Fatalf("WrapBoundary() should unwrap base error")
			}
		})
	}
}
