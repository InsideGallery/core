package instance

import "testing"

func TestInstanceIDs(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "instance id is stable and non-empty",
			run: func(t *testing.T) {
				t.Helper()

				first := GetInstanceID()
				second := GetInstanceID()

				if first == "" {
					t.Fatal("instance id is empty")
				}

				if first != second {
					t.Fatalf("instance id changed from %q to %q", first, second)
				}
			},
		},
		{
			name: "short instance id is stable and non-empty",
			run: func(t *testing.T) {
				t.Helper()

				first := GetShortInstanceID()
				second := GetShortInstanceID()

				if first == "" {
					t.Fatal("short instance id is empty")
				}

				if first != second {
					t.Fatalf("short instance id changed from %q to %q", first, second)
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}
