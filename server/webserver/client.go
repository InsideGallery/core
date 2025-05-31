//go:generate mockgen -source=client.go -destination=mocks/client.go
package webserver

import "net/http"

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
