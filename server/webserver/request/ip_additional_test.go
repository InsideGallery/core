package request

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestIsPrivateAddress(t *testing.T) {
	cases := []struct {
		name    string
		address string
		want    bool
		wantErr error
	}{
		{
			name:    "loopback",
			address: "127.0.0.1",
			want:    true,
		},
		{
			name:    "private",
			address: "10.0.0.1",
			want:    true,
		},
		{
			name:    "link local",
			address: "169.254.1.1",
			want:    true,
		},
		{
			name:    "public",
			address: "8.8.8.8",
		},
		{
			name:    "invalid",
			address: "not-an-ip",
			wantErr: ErrAddressIsNotValid,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got, err := IsPrivateAddress(test.address)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("err = %v, want %v", err, test.wantErr)
			}

			if got != test.want {
				t.Fatalf("private = %v, want %v", got, test.want)
			}
		})
	}
}

func TestIPFromRequest(t *testing.T) {
	cases := []struct {
		name          string
		xRealIP       string
		xForwardedFor string
		want          net.IP
	}{
		{
			name:          "first public forwarded address wins",
			xRealIP:       "198.51.100.10",
			xForwardedFor: "10.0.0.1, 203.0.113.7",
			want:          net.ParseIP("203.0.113.7"),
		},
		{
			name:          "real ip is fallback when forwarded addresses are private or invalid",
			xRealIP:       "198.51.100.11",
			xForwardedFor: "10.0.0.1, bad",
			want:          net.ParseIP("198.51.100.11"),
		},
		{
			name:    "real ip without forwarded list",
			xRealIP: "198.51.100.12",
			want:    net.ParseIP("198.51.100.12"),
		},
		{
			name:    "invalid real ip falls back to request ip strings",
			xRealIP: "bad",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/", func(c fiber.Ctx) error {
				ip, err := IPFromRequest(c)
				if err != nil {
					t.Fatalf("ip from request: %v", err)
				}

				if test.want == nil {
					if got := IPStringFromRequest(c); got != c.IP() {
						t.Fatalf("ip strings = %q, want request ip %q", got, c.IP())
					}

					return nil
				}

				if !ip.Equal(test.want) {
					t.Fatalf("ip = %v, want %v", ip, test.want)
				}

				if got := IPStringFromRequest(c); got != test.want.String() {
					t.Fatalf("ip strings = %q, want %q", got, test.want.String())
				}

				return nil
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if test.xRealIP != "" {
				req.Header.Set("X-Real-Ip", test.xRealIP)
			}

			if test.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", test.xForwardedFor)
			}

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app test: %v", err)
			}
			defer resp.Body.Close()
		})
	}
}
