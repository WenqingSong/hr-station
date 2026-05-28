package httpserver

import (
	"net/http"

	"offer-catcher/server/internal/errors"
	"offer-catcher/server/internal/middleware"
	"offer-catcher/server/internal/response"
)

// NewRouter creates and returns the main HTTP handler with all routes registered
// and middleware chain applied.
// Middleware order: Recover → CORS → RequestLogger → handler
func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthzHandler)

	// Apply middleware chain (outermost first)
	// Recover catches panics from all inner handlers
	// CORS handles cross-origin and preflight requests
	// RequestLogger logs method/path/status/duration for each request
	handler := middleware.RequestLogger(mux)
	handler = middleware.CORS(handler)
	handler = middleware.Recover(handler)

	return handler
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, apperror.ErrMethodNotAllowed.HTTPStatus, apperror.ErrMethodNotAllowed.Code, apperror.ErrMethodNotAllowed.Message)
		return
	}

	data := map[string]string{
		"status":  "ok",
		"service": "offer-catcher-server",
	}

	response.OK(w, data)
}
