package aerospike

import "testing"

func TestGetConnectionConfigFromEnvInvalidEnv(t *testing.T) {
	cases := []struct {
		name  string
		key   string
		value string
	}{
		{
			name:  "invalid port",
			key:   "UNIT_AEROSPIKE_PORT",
			value: "bad",
		},
		{
			name:  "invalid connection queue size",
			key:   "UNIT_AEROSPIKE_CONNECTION_QUEUE_SIZE",
			value: "bad",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv(test.key, test.value)

			_, err := GetConnectionConfigFromEnv("UNIT_AEROSPIKE")
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
