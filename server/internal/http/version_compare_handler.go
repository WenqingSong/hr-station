package httpserver

import (
	"net/http"
	"strconv"

	apperror "offer-catcher/server/internal/errors"
	"offer-catcher/server/internal/middleware"
	"offer-catcher/server/internal/model"
	"offer-catcher/server/internal/response"
	"offer-catcher/server/internal/service"
)

// VersionCompareHandler handles version comparison HTTP requests.
type VersionCompareHandler struct {
	versionCompareService *service.VersionCompareService
}

// NewVersionCompareHandler creates a new VersionCompareHandler.
func NewVersionCompareHandler(versionCompareService *service.VersionCompareService) *VersionCompareHandler {
	return &VersionCompareHandler{versionCompareService: versionCompareService}
}

// Compare handles GET /api/resumes/versions/compare?from=1&to=2 — generates mock version comparison.
func (h *VersionCompareHandler) Compare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, apperror.ErrMethodNotAllowed.HTTPStatus, apperror.ErrMethodNotAllowed.Code, apperror.ErrMethodNotAllowed.Message)
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
		return
	}

	// Parse query parameters (K8-5-48-D: from_id/to_id)
	fromStr := r.URL.Query().Get("from_id")
	toStr := r.URL.Query().Get("to_id")

	if fromStr == "" {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "from_id parameter is required")
		return
	}
	if toStr == "" {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "to_id parameter is required")
		return
	}

	fromID, err := strconv.ParseUint(fromStr, 10, 64)
	if err != nil || fromID < 1 {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid from_id")
		return
	}

	toID, err := strconv.ParseUint(toStr, 10, 64)
	if err != nil || toID < 1 {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid to_id")
		return
	}

	if fromID == toID {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "from_id and to_id must be different")
		return
	}

	// Validate and normalize generation_tier
	tier, err := model.NormalizeGenerationTier(r.URL.Query().Get("generation_tier"))
	if err != nil {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, err.Error())
		return
	}

	req := service.VersionCompareRequest{
		FromResumeID:  fromID,
		ToResumeID:    toID,
		GenerationTier: tier,
	}

	resp, err := h.versionCompareService.Compare(r.Context(), userID, req)
	if err != nil {
		handleVersionCompareError(w, err)
		return
	}

	response.OK(w, resp)
}

// handleVersionCompareError maps version compare service errors to HTTP responses using errors.As.
func handleVersionCompareError(w http.ResponseWriter, err error) {
	if ae := apperror.AsAppError(err); ae != nil {
		response.Error(w, ae.HTTPStatus, ae.Code, ae.Message)
		return
	}
	response.Error(w, apperror.ErrInternal.HTTPStatus, apperror.ErrInternal.Code, apperror.ErrInternal.Message)
}
