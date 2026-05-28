package httpserver

import (
	"net/http"

	apperror "offer-catcher/server/internal/errors"
	"offer-catcher/server/internal/health"
	"offer-catcher/server/internal/middleware"
	"offer-catcher/server/internal/response"
)

// RouterDeps holds all dependencies needed by the router.
type RouterDeps struct {
	HealthService  *health.Service
	AuthHandler    *AuthHandler
	AuthMiddleware func(http.Handler) http.Handler
}

// NewRouter creates and returns the main HTTP handler with all routes registered
// and middleware chain applied.
// Middleware order: Recover → CORS → RequestLogger → handler
func NewRouter(deps RouterDeps) http.Handler {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		healthzHandler(w, r, deps.HealthService)
	})

	// Auth routes (public)
	mux.HandleFunc("/api/auth/register", deps.AuthHandler.Register)
	mux.HandleFunc("/api/auth/login", deps.AuthHandler.Login)

	// Protected auth routes
	mux.Handle("/api/auth/me", deps.AuthMiddleware(http.HandlerFunc(deps.AuthHandler.Me)))

	// Apply middleware chain (outermost first)
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
