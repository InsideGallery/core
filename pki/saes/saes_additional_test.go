package saes

import "testing"

func TestSAESAdditional(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "kind and binary restore",
			run: func(t *testing.T) {
				t.Helper()

				cipher, err := NewSAES()
				if err != nil {
					t.Fatalf("new saes: %v", err)
				}

				if cipher.Kind() != Kind {
					t.Fatalf("kind = %q, want %q", cipher.Kind(), Kind)
				}

				raw, err := cipher.ToBinary()
				if err != nil {
					t.Fatalf("to binary: %v", err)
				}

				if len(raw) != AESSIV64 {
					t.Fatalf("raw length = %d, want %d", len(raw), AESSIV64)
				}

				restored, err := FromBinary(raw)
				if err != nil {
					t.Fatalf("from binary: %v", err)
				}

				fromMethod, err := cipher.FromBinary(raw)
				if err != nil {
					t.Fatalf("method from binary: %v", err)
				}

				if restored.Kind() != Kind {
					t.Fatalf("restored kind = %q, want %q", restored.Kind(), Kind)
				}

				if fromMethod.Kind() != Kind {
					t.Fatalf("method restored kind = %q, want %q", fromMethod.Kind(), Kind)
				}
			},
		},
		{
			name: "invalid restored key returns crypto errors",
			run: func(t *testing.T) {
				t.Helper()

				cipher, err := FromBinary([]byte("bad"))
				if err != nil {
					t.Fatalf("from binary: %v", err)
				}

				if _, err := cipher.Encrypt([]byte("data")); err == nil {
					t.Fatal("expected encrypt error")
				}

				if _, err := cipher.Decrypt([]byte("data")); err == nil {
					t.Fatal("expected decrypt error")
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}
