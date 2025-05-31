package backoff

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/InsideGallery/core/testutils"
)

func TestHTTPTransport_RoundTrip(t *testing.T) {
	tests := []struct {
		name            string
		wantResp        int
		wantErr         error
		backoffType     PoliticType
		expectedCounter int
		delay           time.Duration
		retries         int
	}{
		{
			name:            "success constant backoff",
			wantResp:        http.StatusOK,
			backoffType:     ConstantBackoff,
			expectedCounter: 2,
			delay:           10 * time.Millisecond,
			retries:         2,
		},
		{
			name:            "success exponential backoff",
			wantResp:        http.StatusOK,
			backoffType:     ExponentialBackoff,
			expectedCounter: 2,
			delay:           10 * time.Millisecond,
			retries:         2,
		},
		{
			name:            "success no backoff",
			wantResp:        http.StatusInternalServerError,
			backoffType:     NoBackoff,
			expectedCounter: 1,
			delay:           10 * time.Millisecond,
			retries:         2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			retries := tt.retries
			delay := tt.delay
			counter := 0
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				counter++
				if counter == retries {
					w.WriteHeader(http.StatusOK)
				} else {
					w.WriteHeader(http.StatusInternalServerError)
				}
			}))

			defer testServer.Close()

			req, err := http.NewRequest(http.MethodGet, testServer.URL, nil)
			testutils.Equal(t, err, nil)

			s := NewTransport(delay, retries, tt.backoffType, http.DefaultTransport)
			gotResp, err := s.RoundTrip(req)
			testutils.Equal(t, err, tt.wantErr)
			testutils.Equal(t, gotResp.StatusCode, tt.wantResp)
			testutils.Equal(t, counter, tt.expectedCounter)
		})
	}
}
