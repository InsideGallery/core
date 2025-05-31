package middlewares

import (
	"net/http"
	"strings"
)

// CORSMiddleware implement cors request
func CORSMiddleware(methods ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,X-Requested-With")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
	}
}
