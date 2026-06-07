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
	"offer-catcher/server/internal/service"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const jobIDContextKey contextKey = "job_id"

// RouterDeps holds all dependencies needed by the router.
type RouterDeps struct {
	HealthService             *health.Service
	AuthHandler               *AuthHandler
	AuthMiddleware            func(http.Handler) http.Handler
	AuthService               *service.AuthService
	ResumeHandler             *ResumeHandler
	JobHandler                *JobHandler
	RecommendationHandler     *RecommendationHandler
	MatchHandler              *MatchHandler
	VersionCompareHandler     *VersionCompareHandler
	ResumeOptimizationHandler *ResumeOptimizationHandler
	FrontendURL               string
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
	mux.Handle("/api/auth/logout", deps.AuthMiddleware(http.HandlerFunc(deps.AuthHandler.Logout)))

	// Resume routes (all require auth)
	// K8-5-48-F: /api/resumes dispatches GET (list) / POST (create) via ResumesDispatcher
	mux.Handle("/api/resumes", deps.AuthMiddleware(http.HandlerFunc(deps.ResumeHandler.ResumesDispatcher)))
	// K8-5-48-F: Fixed paths registered before wildcard /api/resumes/
	mux.Handle("/api/resumes/check-name", deps.AuthMiddleware(http.HandlerFunc(deps.ResumeHandler.CheckName))) // K8-5-49-B
	mux.Handle("/api/resumes/current", deps.AuthMiddleware(http.HandlerFunc(deps.ResumeHandler.Current)))
	mux.Handle("/api/resumes/detail", deps.AuthMiddleware(http.HandlerFunc(deps.ResumeHandler.Detail)))
	mux.Handle("/api/resumes/versions", deps.AuthMiddleware(http.HandlerFunc(deps.ResumeHandler.Versions)))
	mux.Handle("/api/resumes/versions/compare", deps.AuthMiddleware(http.HandlerFunc(deps.VersionCompareHandler.Compare)))
	mux.Handle("/api/resumes/compare", deps.AuthMiddleware(http.HandlerFunc(deps.VersionCompareHandler.Compare))) // K8-5-48-D
	mux.Handle("/api/resumes/versions/", deps.AuthMiddleware(http.HandlerFunc(deps.ResumeHandler.VersionDetail)))
	mux.Handle("/api/resumes/optimize", deps.AuthMiddleware(http.HandlerFunc(deps.ResumeOptimizationHandler.Optimize)))
	// K8-5-48-F: Wildcard for /api/resumes/{id} and /api/resumes/{id}/set-current
	mux.Handle("/api/resumes/", deps.AuthMiddleware(http.HandlerFunc(deps.ResumeHandler.ResumeByIDDispatcher)))

	// Job routes — public list / facets / detail, protected recommend / match
	// NOTE: more specific patterns must be registered before the wildcard /api/jobs/
	mux.HandleFunc("/api/jobs", deps.JobHandler.List)                                                              // public
	mux.HandleFunc("/api/jobs/facets", deps.JobHandler.Facets)                                                     // public
	mux.Handle("/api/jobs/recommend", deps.AuthMiddleware(http.HandlerFunc(deps.RecommendationHandler.Recommend))) // protected
	mux.Handle("/api/jobs/recommend/latest", deps.AuthMiddleware(http.HandlerFunc(deps.RecommendationHandler.Latest))) // K9-1-U-C
	mux.Handle("/api/jobs/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dispatchJobWildcard(w, r, deps)
	}))

	// Swagger / OpenAPI docs
	mux.HandleFunc("/swagger/doc.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.ServeFile(w, r, "docs/swagger.json")
	})
	mux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(swaggerUIHTML))
	})

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

// dispatchJobWildcard dispatches /api/jobs/ sub-routes.
// Detail (/api/jobs/{id}) is public.
// Match (/api/jobs/{id}/match) requires auth — enforced internally here.
func dispatchJobWildcard(w http.ResponseWriter, r *http.Request, deps RouterDeps) {
	path := strings.TrimPrefix(r.URL.Path, "/api/jobs/")
	if path == "" {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "job id is required")
		return
	}

	// Check if path matches /api/jobs/{id}/match/latest (protected)
	parts := strings.SplitN(path, "/", 3)
	if len(parts) >= 3 && parts[1] == "match" && parts[2] == "latest" {
		jobID, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil || jobID == 0 {
			response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid job id")
			return
		}
		// Inline auth
		authHeader := r.Header.Get("Authorization")
		parts2 := strings.SplitN(authHeader, " ", 2)
		if len(parts2) != 2 || !strings.EqualFold(parts2[0], "Bearer") {
			response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
			return
		}
		userID, sessionID, err := deps.AuthService.ParseToken(parts2[1])
		if err != nil {
			response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
			return
		}
		if err := deps.AuthService.ValidateSession(r.Context(), userID, sessionID); err != nil {
			log.Printf("[auth] session invalid in match latest dispatch: %v", err)
			response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, middleware.UserIDContextKey, userID)
		ctx = context.WithValue(ctx, jobIDContextKey, jobID)
		deps.MatchHandler.Latest(w, r.WithContext(ctx))
		return
	}

	// Check if path matches /api/jobs/{id}/match (protected)
	parts = strings.SplitN(path, "/", 2)
	if len(parts) == 2 && parts[1] == "match" {
		jobID, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil || jobID == 0 {
			response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid job id")
			return
		}

		// Inline auth: parse token, validate session, inject user_id
		authHeader := r.Header.Get("Authorization")
		parts2 := strings.SplitN(authHeader, " ", 2)
		if len(parts2) != 2 || !strings.EqualFold(parts2[0], "Bearer") {
			response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
			return
		}
		userID, sessionID, err := deps.AuthService.ParseToken(parts2[1])
		if err != nil {
			response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
			return
		}
		if err := deps.AuthService.ValidateSession(r.Context(), userID, sessionID); err != nil {
			log.Printf("[auth] session invalid in match dispatch: %v", err)
			response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
			return
		}

		// Inject userID + jobID into context for MatchHandler
		ctx := r.Context()
		ctx = context.WithValue(ctx, middleware.UserIDContextKey, userID)
		ctx = context.WithValue(ctx, jobIDContextKey, jobID)
		deps.MatchHandler.Analyze(w, r.WithContext(ctx))
		return
	}

	// Default: treat as public job detail request: /api/jobs/{id}
	deps.JobHandler.Detail(w, r)
}

// swaggerUIHTML is a self-contained Swagger UI page served at /swagger/index.html.
const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1"/>
  <title>Offer Catcher API — Swagger</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css"/>
  <style>html{box-sizing:border-box}*,*:before,*:after{box-sizing:inherit}body{margin:0;background:#fafafa}</style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js" crossorigin></script>
  <script>
    SwaggerUIBundle({ url: "/swagger/doc.json", dom_id: "#swagger-ui", deepLinking: true, presets: [SwaggerUIBundle.presets.apis], layout: "BaseLayout" });
  </script>
</body>
</html>`

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
