package embedded

import "testing"

func TestEmbeddedResources(t *testing.T) {
	cases := []struct {
		name string
		read func(name string) ([]byte, error)
	}{
		{
			name: "fs reads embedded resource",
			read: FS().ReadFile,
		},
		{
			name: "deprecated getter reads embedded resource",
			read: GetFS().ReadFile,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			data, err := test.read("resources/tests_for_names.csv")
			if err != nil {
				t.Fatalf("read resource: %v", err)
			}

			if len(data) == 0 {
				t.Fatal("resource is empty")
			}
		})
	}
}
