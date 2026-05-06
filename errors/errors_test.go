package errors

import (
	nativeErrors "errors"
	"testing"
)

var (
	errA = New("error A")
	errB = New("error B")
	errC = New("error C")
	errD = New("error D")
)

type customError struct {
	Code int
	Msg  string
}

func (c *customError) Error() string {
	return c.Msg
}

type anotherCustomError struct {
	Detail string
}

func (a *anotherCustomError) Error() string {
	return a.Detail
}

func TestCombine(t *testing.T) {
	cases := []struct {
		name      string
		input     []error
		wantNil   bool
		wantMsg   string
	}{
		{
			name:    "no arguments returns nil",
			input:   []error{},
			wantNil: true,
		},
		{
			name:    "nil slice returns nil",
			input:   nil,
			wantNil: true,
		},
		{
			name:    "single nil error returns nil",
			input:   []error{nil},
			wantNil: true,
		},
		{
			name:    "multiple nil errors returns nil",
			input:   []error{nil, nil, nil},
			wantNil: true,
		},
		{
			name:    "single non-nil error returned as-is",
			input:   []error{errA},
			wantNil: false,
			wantMsg: "error A",
		},
		{
			name:    "two distinct errors combined",
			input:   []error{errA, errB},
			wantNil: false,
			wantMsg: "error A: error B",
		},
		{
			name:    "three distinct errors combined",
			input:   []error{errA, errB, errC},
			wantNil: false,
			wantMsg: "error A: error B: error C",
		},
		{
			name:    "nil among non-nil errors skipped",
			input:   []error{nil, errA, nil, errB, nil},
			wantNil: false,
			wantMsg: "error A: error B",
		},
		{
			name:    "same error twice deduplicates by message",
			input:   []error{errA, New("error A")},
			wantNil: false,
			wantMsg: "error A",
		},
		{
			name:    "same error instance twice deduplicates",
			input:   []error{errA, errA},
			wantNil: false,
			wantMsg: "error A",
		},
		{
			name:    "four errors combined left-associatively",
			input:   []error{errA, errB, errC, errD},
			wantNil: false,
			wantMsg: "error A: error B: error C: error D",
		},
		{
			name:    "only first non-nil survives when rest are nil",
			input:   []error{nil, nil, errC, nil, nil},
			wantNil: false,
			wantMsg: "error C",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Combine(tc.input...)
			if tc.wantNil {
				if got != nil {
					t.Fatalf("expected nil, got %v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("expected non-nil error, got nil")
			}
			if got.Error() != tc.wantMsg {
				t.Fatalf("expected %q, got %q", tc.wantMsg, got.Error())
			}
		})
	}
}

func TestWrap(t *testing.T) {
	cases := []struct {
		name    string
		cause   error
		effect  error
		wantNil bool
		wantMsg string
	}{
		{
			name:    "both nil returns nil",
			cause:   nil,
			effect:  nil,
			wantNil: true,
		},
		{
			name:    "nil cause returns effect",
			cause:   nil,
			effect:  errB,
			wantNil: false,
			wantMsg: "error B",
		},
		{
			name:    "nil effect returns cause",
			cause:   errA,
			effect:  nil,
			wantNil: false,
			wantMsg: "error A",
		},
		{
			name:    "same message deduplicates to cause",
			cause:   errA,
			effect:  New("error A"),
			wantNil: false,
			wantMsg: "error A",
		},
		{
			name:    "same instance deduplicates to cause",
			cause:   errA,
			effect:  errA,
			wantNil: false,
			wantMsg: "error A",
		},
		{
			name:    "distinct errors produce MultipleError",
			cause:   errA,
			effect:  errB,
			wantNil: false,
			wantMsg: "error A: error B",
		},
		{
			name:    "wrapping already-wrapped error nests further",
			cause:   Wrap(errA, errB),
			effect:  errC,
			wantNil: false,
			wantMsg: "error A: error B: error C",
		},
		{
			name:    "deeply nested wrap",
			cause:   Wrap(Wrap(errA, errB), errC),
			effect:  errD,
			wantNil: false,
			wantMsg: "error A: error B: error C: error D",
		},
		{
			name:    "custom error types wrapped",
			cause:   &customError{Code: 1, Msg: "custom1"},
			effect:  &customError{Code: 2, Msg: "custom2"},
			wantNil: false,
			wantMsg: "custom1: custom2",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Wrap(tc.cause, tc.effect)
			if tc.wantNil {
				if got != nil {
					t.Fatalf("expected nil, got %v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("expected non-nil error, got nil")
			}
			if got.Error() != tc.wantMsg {
				t.Fatalf("expected %q, got %q", tc.wantMsg, got.Error())
			}
		})
	}
}

func TestWrapf(t *testing.T) {
	cases := []struct {
		name    string
		err     error
		format  string
		args    []any
		wantNil bool
		wantMsg string
	}{
		{
			name:    "nil error returns nil",
			err:     nil,
			format:  "something %s",
			args:    []any{"bad"},
			wantNil: true,
		},
		{
			name:    "simple format string",
			err:     errA,
			format:  "wrap: %s",
			args:    []any{"detail"},
			wantNil: false,
			wantMsg: "error A: wrap: detail",
		},
		{
			name:    "format with integer arg",
			err:     errB,
			format:  "code %d",
			args:    []any{42},
			wantNil: false,
			wantMsg: "error B: code 42",
		},
		{
			name:    "format with no args",
			err:     errC,
			format:  "additional context",
			args:    nil,
			wantNil: false,
			wantMsg: "error C: additional context",
		},
		{
			name:    "format with multiple args",
			err:     errA,
			format:  "%s failed with %d retries at %s",
			args:    []any{"op", 3, "server1"},
			wantNil: false,
			wantMsg: "error A: op failed with 3 retries at server1",
		},
		{
			name:    "wrapping custom error type",
			err:     &customError{Code: 500, Msg: "internal"},
			format:  "handler %s",
			args:    []any{"index"},
			wantNil: false,
			wantMsg: "internal: handler index",
		},
		{
			name:    "nil error with empty format returns nil",
			err:     nil,
			format:  "",
			args:    nil,
			wantNil: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Wrapf(tc.err, tc.format, tc.args...)
			if tc.wantNil {
				if got != nil {
					t.Fatalf("expected nil, got %v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("expected non-nil error, got nil")
			}
			if got.Error() != tc.wantMsg {
				t.Fatalf("expected %q, got %q", tc.wantMsg, got.Error())
			}
		})
	}
}

func TestWrapfReturnsPointerMultipleError(t *testing.T) {
	cases := []struct {
		name string
		err  error
	}{
		{
			name: "basic wrapf returns pointer type",
			err:  errA,
		},
		{
			name: "custom error wrapf returns pointer type",
			err:  &customError{Code: 1, Msg: "x"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Wrapf(tc.err, "context")
			if got == nil {
				t.Fatal("expected non-nil")
			}
			if _, ok := got.(*MultipleError); !ok {
				t.Fatalf("expected *MultipleError, got %T", got)
			}
		})
	}
}

func TestMultipleErrorError(t *testing.T) {
	cases := []struct {
		name    string
		me      MultipleError
		wantMsg string
	}{
		{
			name:    "simple cause and effect",
			me:      MultipleError{Cause: errA, Effect: errB},
			wantMsg: "error A: error B",
		},
		{
			name:    "nested cause",
			me:      MultipleError{Cause: MultipleError{Cause: errA, Effect: errB}, Effect: errC},
			wantMsg: "error A: error B: error C",
		},
		{
			name:    "custom error types",
			me:      MultipleError{Cause: &customError{Code: 1, Msg: "c1"}, Effect: &customError{Code: 2, Msg: "c2"}},
			wantMsg: "c1: c2",
		},
		{
			name:    "deeply nested cause chain",
			me:      MultipleError{Cause: MultipleError{Cause: MultipleError{Cause: errA, Effect: errB}, Effect: errC}, Effect: errD},
			wantMsg: "error A: error B: error C: error D",
		},
		{
			name:    "same message cause and effect",
			me:      MultipleError{Cause: New("same"), Effect: New("same")},
			wantMsg: "same: same",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.me.Error()
			if got != tc.wantMsg {
				t.Fatalf("expected %q, got %q", tc.wantMsg, got)
			}
		})
	}
}

func TestMultipleErrorUnwrap(t *testing.T) {
	cases := []struct {
		name      string
		me        MultipleError
		wantCause error
	}{
		{
			name:      "unwrap returns cause",
			me:        MultipleError{Cause: errA, Effect: errB},
			wantCause: errA,
		},
		{
			name:      "unwrap nested returns inner MultipleError",
			me:        MultipleError{Cause: MultipleError{Cause: errA, Effect: errB}, Effect: errC},
			wantCause: MultipleError{Cause: errA, Effect: errB},
		},
		{
			name:      "unwrap custom error cause",
			me:        MultipleError{Cause: &customError{Code: 1, Msg: "x"}, Effect: errA},
			wantCause: &customError{Code: 1, Msg: "x"},
		},
		{
			name:      "unwrap twice reaches root",
			me:        MultipleError{Cause: MultipleError{Cause: errA, Effect: errB}, Effect: errC},
			wantCause: MultipleError{Cause: errA, Effect: errB},
		},
		{
			name:      "unwrap via nativeErrors.Unwrap",
			me:        MultipleError{Cause: errA, Effect: errB},
			wantCause: errA,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.me.Unwrap()
			if got == nil {
				t.Fatal("expected non-nil unwrap result")
			}
			if got.Error() != tc.wantCause.Error() {
				t.Fatalf("expected unwrap %q, got %q", tc.wantCause.Error(), got.Error())
			}
		})
	}
}

func TestMultipleErrorUnwrapChain(t *testing.T) {
	cases := []struct {
		name  string
		err   error
		depth int
		want  string
	}{
		{
			name:  "single wrap unwrap once reaches base",
			err:   Wrap(errA, errB),
			depth: 1,
			want:  "error A",
		},
		{
			name:  "double wrap unwrap once",
			err:   Wrap(Wrap(errA, errB), errC),
			depth: 1,
			want:  "error A: error B",
		},
		{
			name:  "double wrap unwrap twice reaches base",
			err:   Wrap(Wrap(errA, errB), errC),
			depth: 2,
			want:  "error A",
		},
		{
			name:  "triple wrap unwrap three times reaches base",
			err:   Wrap(Wrap(Wrap(errA, errB), errC), errD),
			depth: 3,
			want:  "error A",
		},
		{
			name:  "triple wrap unwrap once",
			err:   Wrap(Wrap(Wrap(errA, errB), errC), errD),
			depth: 1,
			want:  "error A: error B: error C",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cur := tc.err
			for i := 0; i < tc.depth; i++ {
				cur = nativeErrors.Unwrap(cur)
				if cur == nil {
					t.Fatalf("unwrap returned nil at depth %d", i+1)
				}
			}
			if cur.Error() != tc.want {
				t.Fatalf("at depth %d: expected %q, got %q", tc.depth, tc.want, cur.Error())
			}
		})
	}
}

func TestMultipleErrorIs(t *testing.T) {
	wrapped := Wrap(errA, errB)
	deepWrapped := Wrap(Wrap(errA, errB), errC)
	tripleWrapped := Wrap(Wrap(Wrap(errA, errB), errC), errD)

	cases := []struct {
		name   string
		err    error
		target error
		want   bool
	}{
		{
			name:   "matches cause directly",
			err:    wrapped,
			target: errA,
			want:   true,
		},
		{
			name:   "matches effect directly",
			err:    wrapped,
			target: errB,
			want:   true,
		},
		{
			name:   "does not match unrelated error",
			err:    wrapped,
			target: errC,
			want:   false,
		},
		{
			name:   "deeply nested matches root cause",
			err:    deepWrapped,
			target: errA,
			want:   true,
		},
		{
			name:   "deeply nested matches middle effect",
			err:    deepWrapped,
			target: errB,
			want:   true,
		},
		{
			name:   "deeply nested matches outer effect",
			err:    deepWrapped,
			target: errC,
			want:   true,
		},
		{
			name:   "deeply nested does not match unrelated",
			err:    deepWrapped,
			target: errD,
			want:   false,
		},
		{
			name:   "triple nested matches all levels",
			err:    tripleWrapped,
			target: errA,
			want:   true,
		},
		{
			name:   "triple nested matches errD",
			err:    tripleWrapped,
			target: errD,
			want:   true,
		},
		{
			name:   "matches itself",
			err:    wrapped,
			target: wrapped,
			want:   true,
		},
		{
			name:   "error matches equivalent MultipleError",
			err:    Wrap(errA, errB),
			target: Wrap(errA, errB),
			want:   true,
		},
		{
			name:   "nil target with nativeErrors.Is",
			err:    wrapped,
			target: nil,
			want:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := nativeErrors.Is(tc.err, tc.target)
			if got != tc.want {
				t.Fatalf("errors.Is(%v, %v) = %v, want %v", tc.err, tc.target, got, tc.want)
			}
		})
	}
}

func TestMultipleErrorAs(t *testing.T) {
	ce := &customError{Code: 42, Msg: "custom"}

	cases := []struct {
		name      string
		me        MultipleError
		wantMatch bool
	}{
		{
			name:      "As returns true when cause is non-nil",
			me:        MultipleError{Cause: errA, Effect: errB},
			wantMatch: true,
		},
		{
			name:      "As returns true when cause is custom error",
			me:        MultipleError{Cause: ce, Effect: errA},
			wantMatch: true,
		},
		{
			name:      "As returns true with nested MultipleError cause",
			me:        MultipleError{Cause: MultipleError{Cause: errA, Effect: errB}, Effect: errC},
			wantMatch: true,
		},
		{
			name:      "As returns true when effect is custom error",
			me:        MultipleError{Cause: errA, Effect: ce},
			wantMatch: true,
		},
		{
			name:      "As returns true with deeply nested cause chain",
			me:        MultipleError{Cause: MultipleError{Cause: MultipleError{Cause: errA, Effect: errB}, Effect: errC}, Effect: errD},
			wantMatch: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var target *customError
			got := tc.me.As(&target)
			if got != tc.wantMatch {
				t.Fatalf("MultipleError.As() = %v, want %v", got, tc.wantMatch)
			}
		})
	}
}

func TestNativeErrorsAsCustomError(t *testing.T) {
	ce := &customError{Code: 42, Msg: "custom"}

	cases := []struct {
		name      string
		err       error
		wantMatch bool
		wantCode  int
	}{
		{
			name:      "direct custom error matches and populates target",
			err:       ce,
			wantMatch: true,
			wantCode:  42,
		},
		{
			name:      "plain error does not match custom error",
			err:       errA,
			wantMatch: false,
		},
		{
			name:      "nil does not match custom error",
			err:       nil,
			wantMatch: false,
		},
		{
			name:      "custom error wrapped in another custom error",
			err:       &customError{Code: 99, Msg: "other"},
			wantMatch: true,
			wantCode:  99,
		},
		{
			name:      "empty message custom error matches",
			err:       &customError{Code: 0, Msg: ""},
			wantMatch: true,
			wantCode:  0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var target *customError
			got := nativeErrors.As(tc.err, &target)
			if got != tc.wantMatch {
				t.Fatalf("nativeErrors.As = %v, want %v", got, tc.wantMatch)
			}
			if got && target.Code != tc.wantCode {
				t.Fatalf("expected Code %d, got %d", tc.wantCode, target.Code)
			}
		})
	}
}

func TestNativeErrorsAsMultipleErrorPointer(t *testing.T) {
	cases := []struct {
		name      string
		err       error
		wantMatch bool
	}{
		{
			name:      "Wrapf result is *MultipleError and matches",
			err:       Wrapf(errA, "ctx"),
			wantMatch: true,
		},
		{
			name:      "plain error does not match *MultipleError",
			err:       errA,
			wantMatch: false,
		},
		{
			name:      "nil does not match *MultipleError",
			err:       nil,
			wantMatch: false,
		},
		{
			name:      "nested Wrapf matches *MultipleError",
			err:       Wrapf(Wrapf(errA, "inner"), "outer"),
			wantMatch: true,
		},
		{
			name:      "Wrapf with custom error matches *MultipleError",
			err:       Wrapf(&customError{Code: 1, Msg: "x"}, "ctx"),
			wantMatch: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var target *MultipleError
			got := nativeErrors.As(tc.err, &target)
			if got != tc.wantMatch {
				t.Fatalf("nativeErrors.As(*MultipleError) = %v, want %v", got, tc.wantMatch)
			}
		})
	}
}

func TestMultipleErrorIsDirectMethod(t *testing.T) {
	me := MultipleError{Cause: errA, Effect: errB}
	nested := MultipleError{Cause: MultipleError{Cause: errA, Effect: errB}, Effect: errC}

	cases := []struct {
		name   string
		me     MultipleError
		target error
		want   bool
	}{
		{
			name:   "Is matches cause",
			me:     me,
			target: errA,
			want:   true,
		},
		{
			name:   "Is matches effect",
			me:     me,
			target: errB,
			want:   true,
		},
		{
			name:   "Is does not match unrelated",
			me:     me,
			target: errC,
			want:   false,
		},
		{
			name:   "Is on nested matches deep cause",
			me:     nested,
			target: errA,
			want:   true,
		},
		{
			name:   "Is on nested matches outer effect",
			me:     nested,
			target: errC,
			want:   true,
		},
		{
			name:   "Is does not match nil",
			me:     me,
			target: nil,
			want:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.me.Is(tc.target)
			if got != tc.want {
				t.Fatalf("MultipleError.Is(%v) = %v, want %v", tc.target, got, tc.want)
			}
		})
	}
}

func TestMultipleErrorAsDirectMethod(t *testing.T) {
	ce := &customError{Code: 99, Msg: "direct"}
	ace := &anotherCustomError{Detail: "direct-another"}

	cases := []struct {
		name      string
		me        MultipleError
		wantMatch bool
		verify    func(t *testing.T, me MultipleError)
	}{
		{
			name:      "As finds custom error in cause via nativeErrors.As",
			me:        MultipleError{Cause: ce, Effect: errA},
			wantMatch: true,
			verify: func(t *testing.T, me MultipleError) {
				var target *customError
				if !nativeErrors.As(me.Cause, &target) {
					t.Fatal("expected As to match cause")
				}
				if target.Code != 99 {
					t.Fatalf("expected Code 99, got %d", target.Code)
				}
			},
		},
		{
			name:      "As finds anotherCustomError in effect via nativeErrors.As",
			me:        MultipleError{Cause: errA, Effect: ace},
			wantMatch: true,
			verify: func(t *testing.T, me MultipleError) {
				var target *anotherCustomError
				if !nativeErrors.As(me.Effect, &target) {
					t.Fatal("expected As to match effect")
				}
				if target.Detail != "direct-another" {
					t.Fatalf("expected Detail 'direct-another', got %q", target.Detail)
				}
			},
		},
		{
			name:      "As no match in cause for custom error",
			me:        MultipleError{Cause: errA, Effect: errB},
			wantMatch: false,
			verify: func(t *testing.T, me MultipleError) {
				var target *customError
				if nativeErrors.As(me.Cause, &target) {
					t.Fatal("expected As to not match cause")
				}
			},
		},
		{
			name:      "As finds nested custom error in cause chain",
			me:        MultipleError{Cause: MultipleError{Cause: ce, Effect: errA}, Effect: errB},
			wantMatch: true,
			verify: func(t *testing.T, me MultipleError) {
				var target *customError
				if !nativeErrors.As(me.Cause, &target) {
					t.Fatal("expected As to match nested cause")
				}
			},
		},
		{
			name:      "As finds MultipleError type in cause",
			me:        MultipleError{Cause: MultipleError{Cause: errA, Effect: errB}, Effect: errC},
			wantMatch: true,
			verify: func(t *testing.T, me MultipleError) {
				var target MultipleError
				if !nativeErrors.As(me.Cause, &target) {
					t.Fatal("expected As to match MultipleError in cause")
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.verify(t, tc.me)
		})
	}
}

func TestNew(t *testing.T) {
	cases := []struct {
		name string
		text string
	}{
		{
			name: "simple message",
			text: "simple error",
		},
		{
			name: "empty string",
			text: "",
		},
		{
			name: "message with special characters",
			text: "error: something went wrong! @#$%",
		},
		{
			name: "message with newlines",
			text: "line1\nline2",
		},
		{
			name: "very long message",
			text: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := New(tc.text)
			if err == nil {
				t.Fatal("expected non-nil error")
			}
			if err.Error() != tc.text {
				t.Fatalf("expected %q, got %q", tc.text, err.Error())
			}
		})
	}
}

func TestWrapReturnTypes(t *testing.T) {
	cases := []struct {
		name           string
		cause          error
		effect         error
		wantMultiple   bool
		wantIdentity   bool
	}{
		{
			name:         "distinct errors return MultipleError value",
			cause:        errA,
			effect:       errB,
			wantMultiple: true,
		},
		{
			name:         "nil cause returns effect identity",
			cause:        nil,
			effect:       errA,
			wantIdentity: true,
		},
		{
			name:         "nil effect returns cause identity",
			cause:        errA,
			effect:       nil,
			wantIdentity: true,
		},
		{
			name:         "same message returns cause identity",
			cause:        errA,
			effect:       New("error A"),
			wantIdentity: true,
		},
		{
			name:         "nested wrap produces MultipleError",
			cause:        Wrap(errA, errB),
			effect:       errC,
			wantMultiple: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Wrap(tc.cause, tc.effect)
			if tc.wantMultiple {
				if _, ok := got.(MultipleError); !ok {
					t.Fatalf("expected MultipleError, got %T", got)
				}
			}
			if tc.wantIdentity {
				if got == nil {
					return
				}
				if _, ok := got.(MultipleError); ok {
					t.Fatalf("did not expect MultipleError for identity case, got %T", got)
				}
			}
		})
	}
}

func TestCombinePreservesIsChain(t *testing.T) {
	cases := []struct {
		name   string
		input  []error
		target error
		want   bool
	}{
		{
			name:   "combined errors is-check for first",
			input:  []error{errA, errB, errC},
			target: errA,
			want:   true,
		},
		{
			name:   "combined errors is-check for middle",
			input:  []error{errA, errB, errC},
			target: errB,
			want:   true,
		},
		{
			name:   "combined errors is-check for last",
			input:  []error{errA, errB, errC},
			target: errC,
			want:   true,
		},
		{
			name:   "combined errors is-check for unrelated fails",
			input:  []error{errA, errB},
			target: errD,
			want:   false,
		},
		{
			name:   "single combined error matches itself",
			input:  []error{errA},
			target: errA,
			want:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			combined := Combine(tc.input...)
			got := nativeErrors.Is(combined, tc.target)
			if got != tc.want {
				t.Fatalf("errors.Is(Combine(...), %v) = %v, want %v", tc.target, got, tc.want)
			}
		})
	}
}

func TestWrapfUnwrap(t *testing.T) {
	cases := []struct {
		name      string
		err       error
		format    string
		args      []any
		wantCause string
	}{
		{
			name:      "unwrap wrapf returns original error",
			err:       errA,
			format:    "context %d",
			args:      []any{1},
			wantCause: "error A",
		},
		{
			name:      "unwrap wrapf with custom error",
			err:       &customError{Code: 1, Msg: "orig"},
			format:    "wrapped",
			args:      nil,
			wantCause: "orig",
		},
		{
			name:      "unwrap wrapf with nested error",
			err:       Wrap(errA, errB),
			format:    "outer",
			args:      nil,
			wantCause: "error A: error B",
		},
		{
			name:      "unwrap wrapf preserves Is chain",
			err:       errC,
			format:    "info",
			args:      nil,
			wantCause: "error C",
		},
		{
			name:      "wrapf unwrap matches original via Is",
			err:       errD,
			format:    "extra %s",
			args:      []any{"data"},
			wantCause: "error D",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wrapped := Wrapf(tc.err, tc.format, tc.args...)
			if wrapped == nil {
				t.Fatal("expected non-nil")
			}
			unwrapped := nativeErrors.Unwrap(wrapped)
			if unwrapped == nil {
				t.Fatal("unwrap returned nil")
			}
			if unwrapped.Error() != tc.wantCause {
				t.Fatalf("expected %q, got %q", tc.wantCause, unwrapped.Error())
			}
			if !nativeErrors.Is(wrapped, tc.err) {
				t.Fatalf("expected Is to match original error")
			}
		})
	}
}

func TestWrapSymmetryAndEdgeCases(t *testing.T) {
	cases := []struct {
		name    string
		cause   error
		effect  error
		wantMsg string
	}{
		{
			name:    "wrap A,B differs from wrap B,A",
			cause:   errA,
			effect:  errB,
			wantMsg: "error A: error B",
		},
		{
			name:    "wrap B,A reversed",
			cause:   errB,
			effect:  errA,
			wantMsg: "error B: error A",
		},
		{
			name:    "wrap with empty string error",
			cause:   New(""),
			effect:  New("x"),
			wantMsg: ": x",
		},
		{
			name:    "wrap both empty string errors deduplicates",
			cause:   New(""),
			effect:  New(""),
			wantMsg: "",
		},
		{
			name:    "wrap with colon in message",
			cause:   New("a: b"),
			effect:  New("c"),
			wantMsg: "a: b: c",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Wrap(tc.cause, tc.effect)
			if got == nil {
				t.Fatal("expected non-nil")
			}
			if got.Error() != tc.wantMsg {
				t.Fatalf("expected %q, got %q", tc.wantMsg, got.Error())
			}
		})
	}
}
