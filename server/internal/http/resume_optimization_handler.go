package httpserver

import (
	"encoding/json"
	"net/http"

	apperror "offer-catcher/server/internal/errors"
	"offer-catcher/server/internal/middleware"
	"offer-catcher/server/internal/response"
	"offer-catcher/server/internal/service"
)

// ResumeOptimizationHandler handles resume optimization HTTP requests.
type ResumeOptimizationHandler struct {
	resumeOptimizationService *service.ResumeOptimizationService
}

// NewResumeOptimizationHandler creates a new ResumeOptimizationHandler.
func NewResumeOptimizationHandler(resumeOptimizationService *service.ResumeOptimizationService) *ResumeOptimizationHandler {
	return &ResumeOptimizationHandler{resumeOptimizationService: resumeOptimizationService}
}

// Optimize dispatches POST (generate) and GET (retrieve latest) for /api/resumes/optimize.
func (h *ResumeOptimizationHandler) Optimize(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getLatest(w, r)
		return
	case http.MethodPost:
		h.generate(w, r)
		return
	default:
		response.Error(w, apperror.ErrMethodNotAllowed.HTTPStatus, apperror.ErrMethodNotAllowed.Code, apperror.ErrMethodNotAllowed.Message)
	}
}

// getLatest handles GET /api/resumes/optimize — returns the most recent optimization result.
func (h *ResumeOptimizationHandler) getLatest(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
		return
	}

	resp, err := h.resumeOptimizationService.GetLatestOptimization(r.Context(), userID)
	if err != nil {
		handleResumeOptimizationError(w, err)
		return
	}
	if resp == nil {
		response.Error(w, apperror.ErrNotFound.HTTPStatus, apperror.ErrNotFound.Code, "no previous optimization found")
		return
	}

	response.OK(w, resp)
}

// generate handles POST /api/resumes/optimize — generates and saves a new optimization.
func (h *ResumeOptimizationHandler) generate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
		return
	}

	// Parse request body
	var req service.ResumeOptimizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid request body")
		return
	}

	resp, err := h.resumeOptimizationService.Optimize(r.Context(), userID, req)
	if err != nil {
		handleResumeOptimizationError(w, err)
		return
	}

	response.OK(w, resp)
}

// handleResumeOptimizationError maps resume optimization service errors to HTTP responses using errors.As.
func handleResumeOptimizationError(w http.ResponseWriter, err error) {
	if ae := apperror.AsAppError(err); ae != nil {
		response.Error(w, ae.HTTPStatus, ae.Code, ae.Message)
		return
	}
	response.Error(w, apperror.ErrInternal.HTTPStatus, apperror.ErrInternal.Code, apperror.ErrInternal.Message)
}
