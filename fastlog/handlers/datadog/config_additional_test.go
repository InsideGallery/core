package datadog

import "testing"

func TestGetConfigFromEnvInvalidEnv(t *testing.T) {
	cases := []struct {
		name  string
		key   string
		value string
	}{
		{
			name:  "invalid timeout",
			key:   "DATADOG_TIMEOUT",
			value: "bad",
		},
		{
			name:  "invalid level",
			key:   "DATADOG_LEVEL",
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
