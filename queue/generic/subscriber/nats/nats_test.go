package nats

import (
	"testing"
	"time"

	natsgo "github.com/nats-io/nats.go"

	"github.com/InsideGallery/core/queue/generic/subscriber/interfaces"
	"github.com/InsideGallery/core/queue/nats/client"
)

func TestNATSAdapter(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "message exposes subject header data and reply",
			run: func(t *testing.T) {
				t.Helper()

				msg := &Message{
					Msg: &natsgo.Msg{
						Data:    []byte("payload"),
						Reply:   "reply",
						Subject: "subject",
					},
				}

				msg.SetHeader("key", "value")
				copyMsg := msg.Copy("copy")

				if msg.Subject() != "subject" {
					t.Fatalf("subject = %q, want subject", msg.Subject())
				}

				if !msg.IsReply() {
					t.Fatal("message should be a reply")
				}

				if msg.Header()["key"][0] != "value" {
					t.Fatalf("header = %q, want value", msg.Header()["key"][0])
				}

				if string(interfaces.Data(copyMsg)) != "payload" {
					t.Fatalf("copy data = %q, want payload", interfaces.Data(copyMsg))
				}
			},
		},
		{
			name: "config exposes subscriber settings",
			run: func(t *testing.T) {
				t.Helper()

				config := &Config{
					Config: &client.Config{
						ConcurrentSize:    3,
						MaxConcurrentSize: 9,
						ReadTimeout:       time.Second,
					},
				}

				if config.ConcurrentSize() != 3 {
					t.Fatalf("concurrent size = %d, want 3", config.ConcurrentSize())
				}

				if config.MaxConcurrentSize() != 9 {
					t.Fatalf("max concurrent size = %d, want 9", config.MaxConcurrentSize())
				}

				if config.ReadTimeout() != time.Second {
					t.Fatalf("read timeout = %s, want 1s", config.ReadTimeout())
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}
