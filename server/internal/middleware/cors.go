package middleware

import (
	"net/http"
	"strings"

	"offer-catcher/server/internal/config"
)

// corsConfig holds the resolved CORS configuration.
var corsConfig struct {
	allowedOrigins   map[string]bool
	allowedMethods   string
	allowedHeaders   string
	allowCredentials bool
}

// InitCORS initializes CORS configuration from the provided config.
// Must be called before CORS middleware is used.
func InitCORS(cfg config.CORSConfig) {
	allowedMap := make(map[string]bool, len(cfg.AllowedOrigins))
	for _, origin := range cfg.AllowedOrigins {
		allowedMap[origin] = true
	}
	corsConfig.allowedOrigins = allowedMap
	corsConfig.allowedMethods = cfg.AllowedMethods
	corsConfig.allowedHeaders = cfg.AllowedHeaders
	corsConfig.allowCredentials = cfg.AllowCredentials
}

// maxAge is the duration (in seconds) the browser may cache preflight results.
const maxAge = 86400

// CORS handles Cross-Origin Resource Sharing for development environment.
// It checks the Origin header against the configured allowlist, sets appropriate
// CORS response headers, and handles OPTIONS preflight requests with 204.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if isAllowedOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if corsConfig.allowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			w.Header().Set("Access-Control-Allow-Methods", corsConfig.allowedMethods)
			w.Header().Set("Access-Control-Allow-Headers", corsConfig.allowedHeaders)
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

// isAllowedOrigin checks if the origin is in the configured allowlist or matches
// a local development pattern (localhost/127.0.0.1 on any port).
func isAllowedOrigin(origin string) bool {
	if origin == "" {
		return false
	}
	if corsConfig.allowedOrigins[origin] {
		return true
	}
	// Allow any localhost/127.0.0.1 origin for local development convenience
	if strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:") {
		return true
	}
	// Also allow plain localhost/127.0.0.1 without port (Nginx same-origin proxy)
	if origin == "http://localhost" || origin == "http://127.0.0.1" {
		return true
	}
	return false
}
