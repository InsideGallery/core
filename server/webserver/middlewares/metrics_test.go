package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

type spyMetricRecorder struct {
	distributions []metricCall
	counts        []metricCall
}

type metricCall struct {
	name string
	tags []string
}

func (s *spyMetricRecorder) Count(name string, _ int64, tags []string) error {
	s.counts = append(s.counts, metricCall{name: name, tags: tags})

	return nil
}

func (s *spyMetricRecorder) Distribution(name string, _ float64, tags []string) error {
	s.distributions = append(s.distributions, metricCall{name: name, tags: tags})

	return nil
}

func TestMetrics(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		method         string
		target         string
		route          string
		responseStatus int
		wantStatus     int
		wantCounts     []string
		wantNoCounts   []string
		wantTags       []string
	}{
		{
			name:           "success records request metrics",
			method:         http.MethodPost,
			target:         "/users/42",
			route:          "/users/:id",
			responseStatus: http.StatusOK,
			wantStatus:     http.StatusOK,
			wantCounts:     []string{httpRequestCount},
			wantTags: []string{
				"method:POST",
				"route:/users/:id",
				"status_code:200",
			},
		},
		{
			name:           "server error records error metric",
			method:         http.MethodGet,
			target:         "/fail",
			route:          "/fail",
			responseStatus: http.StatusInternalServerError,
			wantStatus:     http.StatusInternalServerError,
			wantCounts:     []string{httpRequestCount, httpRequestError},
			wantTags: []string{
				"method:GET",
				"route:/fail",
				"status_code:500",
			},
		},
		{
			name:         "not found uses unmatched route without error",
			method:       http.MethodGet,
			target:       "/missing",
			wantStatus:   http.StatusNotFound,
			wantCounts:   []string{httpRequestCount},
			wantNoCounts: []string{httpRequestError},
			wantTags: []string{
				"method:GET",
				"route:unmatched",
				"status_code:404",
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			spy := &spyMetricRecorder{}
			app := fiber.New()
			app.Use(Metrics(spy))

			if test.route != "" {
				app.Add([]string{test.method}, test.route, func(ctx fiber.Ctx) error {
					return ctx.Status(test.responseStatus).SendString("ok")
				})
			}

			req := httptest.NewRequest(test.method, test.target, nil)

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test error: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != test.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, test.wantStatus)
			}

			assertCallName(t, spy.distributions, httpRequestDuration)

			for _, want := range test.wantCounts {
				assertCallName(t, spy.counts, want)
			}

			for _, want := range test.wantNoCounts {
				assertNoCallName(t, spy.counts, want)
			}

			for _, want := range test.wantTags {
				assertTag(t, spy.distributions[0].tags, want)
			}
		})
	}
}

func assertCallName(t *testing.T, calls []metricCall, want string) {
	t.Helper()

	for _, call := range calls {
		if call.name == want {
			return
		}
	}

	t.Errorf("calls %v missing %q", calls, want)
}

func assertNoCallName(t *testing.T, calls []metricCall, want string) {
	t.Helper()

	for _, call := range calls {
		if call.name == want {
			t.Errorf("calls %v include %q", calls, want)

			return
		}
	}
}

func assertTag(t *testing.T, tags []string, want string) {
	t.Helper()

	for _, tag := range tags {
		if tag == want {
			return
		}
	}

	t.Errorf("tags %v missing %q", tags, want)
}
