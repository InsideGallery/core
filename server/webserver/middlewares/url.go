package middlewares

import (
	"net/http"
	"strings"
)

func URLWithoutQuery(r *http.Request) string {
	result := r.URL.Opaque
	if result == "" {
		result = r.URL.EscapedPath()
		if result == "" {
			result = "/"
		}
	} else if strings.HasPrefix(result, "//") {
		result = r.URL.Scheme + ":" + result
	}

	return result
}
