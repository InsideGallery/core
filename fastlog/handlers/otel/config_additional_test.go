package otel

import "testing"

func TestGetConfigFromEnvInvalidEnv(t *testing.T) {
	cases := []struct {
		name  string
		key   string
		value string
	}{
		{
			name:  "invalid level",
			key:   "OTEL_LEVEL",
			value: "bad",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv(test.key, test.value)

			_, err := GetConfigFromEnv()
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
