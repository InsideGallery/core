package honeypot

import "testing"

func TestHoneypotReturnsListenError(t *testing.T) {
	cases := []struct {
		name string
		port string
	}{
		{
			name: "invalid port",
			port: "bad:port",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if err := Honeypot(test.port); err == nil {
				t.Fatal("expected listen error")
			}
		})
	}
}
