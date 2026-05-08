package aerospike

import "testing"

func TestSetup(t *testing.T) {
	cases := []struct {
		name         string
		architecture BufferArchitecture
		wantArch64   bool
		wantArch32   bool
	}{
		{
			name: "disables explicit architecture flags",
			architecture: func() BufferArchitecture {
				arch64Bits := true
				arch32Bits := true

				return BufferArchitecture{
					Arch64Bits: &arch64Bits,
					Arch32Bits: &arch32Bits,
				}
			}(),
		},
		{
			name: "ignores missing dependencies",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			Setup(test.architecture)
			Setup(test.architecture)

			if test.architecture.Arch64Bits != nil && *test.architecture.Arch64Bits != test.wantArch64 {
				t.Fatalf("Arch64Bits = %t, want %t", *test.architecture.Arch64Bits, test.wantArch64)
			}

			if test.architecture.Arch32Bits != nil && *test.architecture.Arch32Bits != test.wantArch32 {
				t.Fatalf("Arch32Bits = %t, want %t", *test.architecture.Arch32Bits, test.wantArch32)
			}
		})
	}
}
