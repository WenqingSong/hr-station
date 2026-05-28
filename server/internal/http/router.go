package httpserver

import (
	"net/http"

	"offer-catcher/server/internal/errors"
	"offer-catcher/server/internal/health"
	"offer-catcher/server/internal/middleware"
	"offer-catcher/server/internal/response"
)

// NewRouter creates and returns the main HTTP handler with all routes registered
// and middleware chain applied.
// Middleware order: Recover → CORS → RequestLogger → handler
func NewRouter(healthService *health.Service) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		healthzHandler(w, r, healthService)
	})

	// Apply middleware chain (outermost first)
	// Recover catches panics from all inner handlers
	// CORS handles cross-origin and preflight requests
	// RequestLogger logs method/path/status/duration for each request
	handler := middleware.RequestLogger(mux)
	handler = middleware.CORS(handler)
	handler = middleware.Recover(handler)

	return handler
}

func healthzHandler(w http.ResponseWriter, r *http.Request, healthService *health.Service) {
	if r.Method != http.MethodGet {
		response.Error(w, apperror.ErrMethodNotAllowed.HTTPStatus, apperror.ErrMethodNotAllowed.Code, apperror.ErrMethodNotAllowed.Message)
		return
	}

	status, healthy := healthService.Check(r.Context())

	if healthy {
		response.OK(w, status)
	} else {
		response.OKWithMessage(w, "degraded", status)
	}
}
