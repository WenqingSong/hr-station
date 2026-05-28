package middleware

import (
	"net/http"
	"strings"
)

// allowedOrigins lists the origins permitted for CORS in the development environment.
var allowedOrigins = map[string]bool{
	"http://localhost:5173": true,
	"http://127.0.0.1:5173": true,
}

// allowedMethods is the list of HTTP methods permitted by CORS.
const allowedMethods = "GET, POST, PUT, PATCH, DELETE, OPTIONS"

// allowedHeaders is the list of request headers permitted by CORS.
const allowedHeaders = "Content-Type, Authorization, X-Requested-With"

// maxAge is the duration (in seconds) the browser may cache preflight results.
const maxAge = 86400

// CORS handles Cross-Origin Resource Sharing for development environment.
// It allows requests from localhost:5173 and 127.0.0.1:5173, supports common
// methods and headers, and handles OPTIONS preflight requests with 204.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
			w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
			w.Header().Set("Access-Control-Max-Age", "86400")
		}

		// Handle preflight requests
		if r.Method == http.MethodOptions && isAllowedOrigin(origin) {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// isAllowedOrigin checks if the origin is in the allowed list or is a local development origin.
// For local development convenience, any localhost or 127.0.0.1 origin on any port is accepted.
func isAllowedOrigin(origin string) bool {
	if allowedOrigins[origin] {
		return true
	}
	// Allow any localhost/127.0.0.1 origin for local development convenience
	if strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:") {
		return true
	}
	return false
}
