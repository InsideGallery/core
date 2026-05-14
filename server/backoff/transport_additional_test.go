package backoff

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestTransportDefaultsAndErrors(t *testing.T) {
	cases := []struct {
		name    string
		backoff PoliticType
	}{
		{name: "constant backoff retries errors", backoff: ConstantBackoff},
		{name: "exponential backoff retries errors", backoff: ExponentialBackoff},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			transport := NewTransport(0, 0, test.backoff, roundTripperFunc(func(*http.Request) (*http.Response, error) {
				return nil, errors.New("temporary")
			}))

			req, err := http.NewRequest(http.MethodGet, "https://inside.test", nil)
			if err != nil {
				t.Fatalf("new request: %v", err)
			}

			if _, err := transport.RoundTrip(req); err == nil {
				t.Fatal("expected transport error")
			}

			if transport.delay != DefaultDelay {
				t.Fatalf("delay = %v, want %v", transport.delay, DefaultDelay)
			}

			if transport.retries != DefaultRetries {
				t.Fatalf("retries = %d, want %d", transport.retries, DefaultRetries)
			}
		})
	}
}

func TestSetupClientBackoffAndMaxDelay(t *testing.T) {
	client := &http.Client{}
	SetupClientBackoff(client, time.Millisecond, 2, ExponentialBackoff)

	if _, ok := client.Transport.(*HTTPTransport); !ok {
		t.Fatalf("transport = %T, want *HTTPTransport", client.Transport)
	}

	transport := &HTTPTransport{}
	if got := transport.getDelay(DefaultMaxInterval); got != DefaultMaxInterval {
		t.Fatalf("delay = %v, want %v", got, DefaultMaxInterval)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
