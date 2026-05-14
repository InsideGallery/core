package honeypot

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

func TestHandleConnection(t *testing.T) {
	cases := []struct {
		name  string
		lines []string
	}{
		{
			name:  "client sends one line",
			lines: []string{"hello"},
		},
		{
			name: "client closes after banner",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			serverConn, clientConn := net.Pipe()
			done := make(chan struct{})

			go func() {
				handleConnection(serverConn)
				close(done)
			}()

			reader := bufio.NewReader(clientConn)
			banner, err := reader.ReadString('\n')
			if err != nil {
				t.Fatalf("read banner: %v", err)
			}

			if !strings.Contains(banner, "SSH-2.0-OpenSSH") {
				t.Fatalf("banner = %q, want fake ssh banner", banner)
			}

			for _, line := range test.lines {
				if _, err := fmt.Fprintln(clientConn, line); err != nil {
					t.Fatalf("write line: %v", err)
				}
			}

			if err := clientConn.Close(); err != nil {
				t.Fatalf("close client: %v", err)
			}

			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("connection handler did not stop")
			}
		})
	}
}
