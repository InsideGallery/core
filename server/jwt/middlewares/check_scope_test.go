//go:build unit
// +build unit

package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/InsideGallery/core/testutils"
)

func Test_parseScope(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		target         string
		requestURI     string
		want           string
		wantStatusCode int
	}{
		{
			name:           "read GET",
			method:         http.MethodGet,
			target:         "/organizations",
			want:           "read:organizations:",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "read GET",
			method:         http.MethodGet,
			target:         "/organizations/some-id",
			want:           "read:organizations:some-id",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "read GET",
			method:         http.MethodGet,
			target:         "/v1/organizations/some-id",
			want:           "read:organizations:some-id",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "read OPTION",
			method:         http.MethodOptions,
			target:         "/organizations",
			want:           "read:organizations:",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "read OPTION",
			method:         http.MethodOptions,
			target:         "/organizations/some-id",
			want:           "read:organizations:some-id",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "read HEAD",
			method:         http.MethodHead,
			target:         "/organizations",
			want:           "read:organizations:",
			wantStatusCode: http.StatusOK,
		},
		// todo: потрібно прописати ендпоінти для ролбеку обʼєктів
		{
			name:           "read HEAD",
			method:         http.MethodHead,
			target:         "/organizations/some-id",
			want:           "read:organizations:some-id",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "write POST",
			method:         http.MethodPost,
			target:         "/organizations",
			want:           "write:organizations:",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "write POST",
			method:         http.MethodPost,
			target:         "/organizations/some-id",
			want:           "write:organizations:some-id",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "write PUT",
			method:         http.MethodPut,
			target:         "/organizations",
			want:           "write:organizations:",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "write PUT",
			method:         http.MethodPut,
			target:         "/organizations/some-id",
			want:           "write:organizations:some-id",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "write PATCH",
			method:         http.MethodPatch,
			target:         "/organizations",
			want:           "write:organizations:",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "write PATCH",
			method:         http.MethodPatch,
			target:         "/organizations/some-id",
			want:           "write:organizations:some-id",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "write DELETE",
			method:         http.MethodDelete,
			target:         "/organizations",
			want:           "write:organizations:",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "write DELETE",
			method:         http.MethodDelete,
			target:         "/organizations/some-id",
			want:           "write:organizations:some-id",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "read GET with query params",
			method:         http.MethodGet,
			target:         "/v1/policies/history",
			requestURI:     "/v1/policies/history?page=1&per-page=100",
			want:           "read:policies:history",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "read GET entity by ID",
			method:         http.MethodGet,
			target:         "/v1/policies/history/:id",
			requestURI:     "/v1/policies/history/64ac106b7cb02d8d34948708",
			want:           "read:policies:history/64ac106b7cb02d8d34948708",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "write PUT entity by ID",
			method:         http.MethodPut,
			target:         "/v1/policies/history/:id",
			requestURI:     "/v1/policies/history/64ac106b7cb02d8d34948708",
			want:           "write:policies:history/64ac106b7cb02d8d34948708",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "write PATCH entity by ID",
			method:         http.MethodPatch,
			target:         "/v1/policies/history/:id",
			requestURI:     "/v1/policies/history/64ac106b7cb02d8d34948708",
			want:           "write:policies:history/64ac106b7cb02d8d34948708",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "write PUT entity by ID",
			method:         http.MethodDelete,
			target:         "/v1/policies/history/:id",
			requestURI:     "/v1/policies/history/64ac106b7cb02d8d34948708",
			want:           "write:policies:history/64ac106b7cb02d8d34948708",
			wantStatusCode: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var triggered bool

			webServer := fiber.New()
			webServer.All(tt.target, func(c *fiber.Ctx) error {
				triggered = true

				if got := parseScope(c); got != tt.want {
					t.Errorf("parseScope() = %v, want %v", got, tt.want)
				}

				return nil
			})

			var req *http.Request

			if tt.requestURI == "" {
				req = httptest.NewRequest(tt.method, tt.target, nil)
			} else {
				req = httptest.NewRequest(tt.method, tt.requestURI, nil)
			}

			resp, err := webServer.Test(req, -1)
			testutils.Equal(t, err, nil)
			testutils.Equal(t, triggered, true)
			testutils.Equal(t, resp.StatusCode, tt.wantStatusCode)
		})
	}
}
