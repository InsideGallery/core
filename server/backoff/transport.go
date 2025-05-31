package backoff

import (
	"net/http"
	"sync"
	"time"
)

type PoliticType int

const (
	NoBackoff PoliticType = iota
	ExponentialBackoff
	ConstantBackoff

	DefaultRetries                = 1
	DefaultMaxInterval            = 60 * time.Second
	DefaultExponentialMultiplayer = 2
	DefaultDelay                  = 250 * time.Millisecond
)

type HTTPTransport struct {
	http.RoundTripper
	retries int
	backoff PoliticType
	delay   time.Duration
	mu      sync.RWMutex
}

func SetupClientBackoff(client *http.Client, delay time.Duration, retries int, backoff PoliticType) {
	client.Transport = NewTransport(delay, retries, backoff, client.Transport)
}

func NewTransport(delay time.Duration, retries int, backoff PoliticType, tripper http.RoundTripper) *HTTPTransport {
	if retries <= 0 {
		retries = DefaultRetries
	}

	if delay <= 0 {
		delay = DefaultDelay
	}

	return &HTTPTransport{
		RoundTripper: tripper,
		backoff:      backoff,
		retries:      retries,
		delay:        delay,
	}
}

func (s *HTTPTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.RoundTripper == nil {
		s.RoundTripper = http.DefaultTransport
	}

	switch s.backoff {
	case ConstantBackoff:
		for i := 0; i < s.retries; i++ {
			req.Close = true

			resp, err = s.RoundTripper.RoundTrip(req)
			if err != nil {
				time.Sleep(s.delay)
				continue
			}

			if resp.StatusCode >= http.StatusBadRequest {
				time.Sleep(s.delay)
				continue
			}

			return resp, err
		}
	case ExponentialBackoff:
		delay := s.delay

		for i := 0; i < s.retries; i++ {
			req.Close = true

			resp, err = s.RoundTripper.RoundTrip(req)
			if err != nil {
				time.Sleep(delay)
				delay = s.getDelay(delay)

				continue
			}

			if resp.StatusCode >= http.StatusBadRequest {
				time.Sleep(delay)
				delay = s.getDelay(delay)

				continue
			}

			return resp, err
		}
	default:
		resp, err = s.RoundTripper.RoundTrip(req)
	}

	return resp, err
}

func (*HTTPTransport) getDelay(delay time.Duration) time.Duration {
	delay *= time.Duration(DefaultExponentialMultiplayer)
	if delay >= DefaultMaxInterval {
		delay = DefaultMaxInterval
	}

	return delay
}
