package utils

import (
	"strings"
	"testing"
)

func TestStringHelpersAdditional(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "ab test handles empty and weighted groups",
			run: func(t *testing.T) {
				t.Helper()

				if got := ABTest([]byte("user"), []byte("salt")); got != 0 {
					t.Fatalf("empty groups = %d, want 0", got)
				}

				if got := ABTest([]byte("user"), []byte("salt"), 2, 3); got >= 5 {
					t.Fatalf("weighted group = %d, want less than 5", got)
				}
			},
		},
		{
			name: "email username",
			run: func(t *testing.T) {
				t.Helper()

				tests := []struct {
					email string
					want  string
				}{
					{email: "user@example.test", want: "user"},
					{email: "@example.test", want: ""},
					{email: "plain", want: "plain"},
				}

				for _, test := range tests {
					if got := EmailUserName(test.email); got != test.want {
						t.Fatalf("username(%q) = %q, want %q", test.email, got, test.want)
					}
				}
			},
		},
		{
			name: "unicode normalization lowercases and trims",
			run: func(t *testing.T) {
				t.Helper()

				if got := NFDLowerString(" É "); got != "e\u0301" {
					t.Fatalf("nfd = %q, want decomposed lower e", got)
				}

				if got := NFKDLowerString(" É "); got != "e\u0301" {
					t.Fatalf("nfkd = %q, want decomposed lower e", got)
				}
			},
		},
		{
			name: "unique id helpers return expected sizes",
			run: func(t *testing.T) {
				t.Helper()

				if got := GetUniqueID(); len(got) != 24 {
					t.Fatalf("unique id length = %d, want 24", len(got))
				}

				shortID, err := GetShortID()
				if err != nil {
					t.Fatalf("short id: %v", err)
				}

				if len(shortID) != 12 {
					t.Fatalf("short id length = %d, want 12", len(shortID))
				}
			},
		},
		{
			name: "safe get returns pointer value or default",
			run: func(t *testing.T) {
				t.Helper()

				value := "set"
				if got := SafeGet(&value, "default"); got != "set" {
					t.Fatalf("safe get pointer = %q, want set", got)
				}

				if got := SafeGet[string](nil, "default"); got != "default" {
					t.Fatalf("safe get nil = %q, want default", got)
				}
			},
		},
		{
			name: "split chunks handles invalid and uneven chunks",
			run: func(t *testing.T) {
				t.Helper()

				if got := SplitByChunks("abc", 0); got != nil {
					t.Fatalf("invalid chunk = %v, want nil", got)
				}

				got := SplitByChunks("abcdefg", 3)
				if strings.Join(got, "|") != "abc|def|g" {
					t.Fatalf("chunks = %v", got)
				}
			},
		},
		{
			name: "rand string handles non-positive size",
			run: func(t *testing.T) {
				t.Helper()

				if got := RandStringBytes(0); got != "" {
					t.Fatalf("rand string = %q, want empty", got)
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}
