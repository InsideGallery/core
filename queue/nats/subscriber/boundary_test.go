package subscriber

import (
	"testing"

	natsgo "github.com/nats-io/nats.go"
)

func TestNewMessageCopiesNATSMessage(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		msg  *natsgo.Msg
		want Message
	}{
		{
			name: "nil message",
			msg:  nil,
			want: Message{},
		},
		{
			name: "copies fields",
			msg: &natsgo.Msg{
				Subject: "events",
				Reply:   "reply",
				Data:    []byte("payload"),
				Header:  natsgo.Header{"x-test": []string{"true"}},
			},
			want: Message{
				Subject: "events",
				Reply:   "reply",
				Data:    []byte("payload"),
				Header:  Headers{"x-test": []string{"true"}},
			},
		},
	}

	for _, test := range cases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := newMessage(test.msg)
			if got.Subject != test.want.Subject {
				t.Fatalf("Subject = %q, want %q", got.Subject, test.want.Subject)
			}

			if got.Reply != test.want.Reply {
				t.Fatalf("Reply = %q, want %q", got.Reply, test.want.Reply)
			}

			if string(got.Data) != string(test.want.Data) {
				t.Fatalf("Data = %q, want %q", got.Data, test.want.Data)
			}

			for key, values := range test.want.Header {
				if got.Header[key][0] != values[0] {
					t.Fatalf("Header[%q] = %q, want %q", key, got.Header[key][0], values[0])
				}
			}
		})
	}
}

func TestMessageSubscriberContract(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		assert func(t *testing.T)
	}{
		{
			name: "subscriber implements message subscriber",
			assert: func(t *testing.T) {
				t.Helper()

				var _ MessageSubscriber = (*Subscriber)(nil)
			},
		},
	}

	for _, test := range cases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			test.assert(t)
		})
	}
}
