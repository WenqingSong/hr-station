package httpserver

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	apperror "offer-catcher/server/internal/errors"
	"offer-catcher/server/internal/health"
	"offer-catcher/server/internal/middleware"
	"offer-catcher/server/internal/response"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const jobIDContextKey contextKey = "job_id"

// RouterDeps holds all dependencies needed by the router.
type RouterDeps struct {
	HealthService              *health.Service
	AuthHandler                *AuthHandler
	AuthMiddleware             func(http.Handler) http.Handler
	ResumeHandler              *ResumeHandler
	JobHandler                 *JobHandler
	RecommendationHandler      *RecommendationHandler
	MatchHandler               *MatchHandler
	VersionCompareHandler      *VersionCompareHandler
	ResumeOptimizationHandler  *ResumeOptimizationHandler
	FrontendURL                string
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

	// Resume routes (all require auth)
	mux.Handle("/api/resumes", deps.AuthMiddleware(http.HandlerFunc(deps.ResumeHandler.Create)))
	mux.Handle("/api/resumes/current", deps.AuthMiddleware(http.HandlerFunc(deps.ResumeHandler.Current)))
	mux.Handle("/api/resumes/detail", deps.AuthMiddleware(http.HandlerFunc(deps.ResumeHandler.Detail)))
	mux.Handle("/api/resumes/versions", deps.AuthMiddleware(http.HandlerFunc(deps.ResumeHandler.Versions)))
	mux.Handle("/api/resumes/versions/compare", deps.AuthMiddleware(http.HandlerFunc(deps.VersionCompareHandler.Compare)))
	mux.Handle("/api/resumes/versions/", deps.AuthMiddleware(http.HandlerFunc(deps.ResumeHandler.VersionDetail)))
	mux.Handle("/api/resumes/optimize", deps.AuthMiddleware(http.HandlerFunc(deps.ResumeOptimizationHandler.Optimize)))

	// Job routes (all require auth)
	mux.Handle("/api/jobs", deps.AuthMiddleware(http.HandlerFunc(deps.JobHandler.List)))
	mux.Handle("/api/jobs/recommend", deps.AuthMiddleware(http.HandlerFunc(deps.RecommendationHandler.Recommend)))
	mux.Handle("/api/jobs/", deps.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dispatchJobRoutes(w, r, deps)
	})))

	// Frontend reverse proxy: forward all non-API requests to Vite dev server
	if deps.FrontendURL != "" {
		frontendProxy := newFrontendProxy(deps.FrontendURL)
		mux.Handle("/", frontendProxy)
	}

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

// dispatchJobRoutes dispatches /api/jobs/ sub-routes based on path pattern.
// - /api/jobs/{id}       → JobHandler.Detail (GET)
// - /api/jobs/{id}/match → MatchHandler.Analyze (POST)
func dispatchJobRoutes(w http.ResponseWriter, r *http.Request, deps RouterDeps) {
	path := strings.TrimPrefix(r.URL.Path, "/api/jobs/")
	if path == "" {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "job id is required")
		return
	}

	// Check if path matches /api/jobs/{id}/match
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 2 && parts[1] == "match" {
		// This is a match analysis request: /api/jobs/{id}/match
		jobID, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil || jobID == 0 {
			response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid job id")
			return
		}
		// Inject jobID into context
		ctx := context.WithValue(r.Context(), jobIDContextKey, jobID)
		deps.MatchHandler.Analyze(w, r.WithContext(ctx))
		return
	}

	// Default: treat as job detail request: /api/jobs/{id}
	deps.JobHandler.Detail(w, r)
}

// newFrontendProxy creates a reverse proxy to the Vite dev server.
// It forwards all requests (including WebSocket for HMR) to the frontend.
func newFrontendProxy(frontendURL string) http.Handler {
	target, err := url.Parse(frontendURL)
	if err != nil {
		log.Printf("[router] invalid FRONTEND_URL %q: %v", frontendURL, err)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "frontend proxy misconfigured", http.StatusInternalServerError)
		})
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host
	}

	// Upgrade handler for Vite HMR WebSocket
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.ToLower(r.Header.Get("Upgrade")) == "websocket" {
			// WebSocket proxy
			wsProxy := httputil.NewSingleHostReverseProxy(target)
			wsProxy.Director = func(req *http.Request) {
				req.URL.Scheme = "ws"
				req.URL.Host = target.Host
				req.Host = target.Host
			}
			wsProxy.ServeHTTP(w, r)
			return
		}
		proxy.ServeHTTP(w, r)
	})
}
