package redis

import "testing"

func TestGetConnectionConfigFromEnvInvalidEnv(t *testing.T) {
	cases := []struct {
		name  string
		key   string
		value string
	}{
		{
			name:  "invalid database",
			key:   "REDIS_DATABASE",
			value: "bad",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv(test.key, test.value)

			_, err := GetConnectionConfigFromEnv()
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
