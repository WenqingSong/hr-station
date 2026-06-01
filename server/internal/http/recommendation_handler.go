package httpserver

import (
	"encoding/json"
	"net/http"

	apperror "offer-catcher/server/internal/errors"
	"offer-catcher/server/internal/middleware"
	"offer-catcher/server/internal/response"
	"offer-catcher/server/internal/service"
)

// RecommendationHandler handles job recommendation HTTP requests.
type RecommendationHandler struct {
	recommendationService *service.RecommendationService
}

// NewRecommendationHandler creates a new RecommendationHandler.
func NewRecommendationHandler(recommendationService *service.RecommendationService) *RecommendationHandler {
	return &RecommendationHandler{recommendationService: recommendationService}
}

// Recommend handles POST /api/jobs/recommend — generates mock job recommendations.
func (h *RecommendationHandler) Recommend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, apperror.ErrMethodNotAllowed.HTTPStatus, apperror.ErrMethodNotAllowed.Code, apperror.ErrMethodNotAllowed.Message)
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
		return
	}

	// Parse request body (optional - can be empty)
	req := service.RecommendationRequest{}
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid request body")
			return
		}
	}

	// Validate resume_version_no
	if req.ResumeVersionNo < 0 {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid resume_version_no")
		return
	}

	// Validate limit
	if req.Limit < 0 || req.Limit > 5 {
		if req.Limit < 0 {
			response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid limit")
			return
		}
		// Limit > 5 will be capped by service
	}

	resp, err := h.recommendationService.Recommend(r.Context(), userID, req)
	if err != nil {
		handleRecommendationError(w, err)
		return
	}

	response.OK(w, resp)
}

// handleRecommendationError maps recommendation service errors to HTTP responses using errors.As.
func handleRecommendationError(w http.ResponseWriter, err error) {
	if ae := apperror.AsAppError(err); ae != nil {
		response.Error(w, ae.HTTPStatus, ae.Code, ae.Message)
		return
	}
	response.Error(w, apperror.ErrInternal.HTTPStatus, apperror.ErrInternal.Code, apperror.ErrInternal.Message)
}
