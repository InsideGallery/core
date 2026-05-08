package bunt

import "testing"

func TestOpen(t *testing.T) {
	cases := []struct {
		name    string
		cfg     *ConnectionConfig
		wantErr bool
	}{
		{
			name: "explicit memory database",
			cfg: &ConnectionConfig{
				Filename: ":memory:",
			},
		},
		{
			name:    "missing config returns error",
			wantErr: true,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			wrapper, err := Open(test.cfg)
			if test.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}

				return
			}

			if err != nil {
				t.Fatalf("open: %v", err)
			}
			defer wrapper.Close()

			if wrapper.DB == nil {
				t.Fatal("buntdb handle is nil")
			}
		})
	}
}

func TestOpenFromEnvCompatibility(t *testing.T) {
	cases := []struct {
		name   string
		prefix string
	}{
		{
			name:   "opens from explicit env prefix",
			prefix: "UNIT_BUNT_OPEN",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv(test.prefix+"_FILENAME", ":memory:")

			wrapper, err := OpenFromEnv(test.prefix)
			if err != nil {
				t.Fatalf("open from env: %v", err)
			}
			defer wrapper.Close()

			if wrapper.DB == nil {
				t.Fatal("buntdb handle is nil")
			}
		})
	}
}
