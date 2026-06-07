package httpserver

import (
	"encoding/json"
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

// Compare dispatches GET (old API) and POST (K9-4-C V2 API) for /api/resumes/compare.
func (h *VersionCompareHandler) Compare(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.CompareV2(w, r)
		return
	}
	if r.Method != http.MethodGet {
		response.Error(w, apperror.ErrMethodNotAllowed.HTTPStatus, apperror.ErrMethodNotAllowed.Code, apperror.ErrMethodNotAllowed.Message)
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
		return
	}

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

// CompareV2 handles POST /api/resumes/compare (K9-4-C V2 API).
func (h *VersionCompareHandler) CompareV2(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
		return
	}

	var req model.CompareRequestV2
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid request body")
			return
		}
	}

	resp, err := h.versionCompareService.CompareV2(r.Context(), userID, req)
	if err != nil {
		handleVersionCompareError(w, err)
		return
	}
	response.OK(w, resp)
}

// LatestV2 handles GET /api/resumes/compare/latest (K9-4-D).
func (h *VersionCompareHandler) LatestV2(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
		return
	}

	fromStr := r.URL.Query().Get("from_resume_id")
	toStr := r.URL.Query().Get("to_resume_id")
	mode := r.URL.Query().Get("comparison_mode")
	targetDir := r.URL.Query().Get("target_direction")
	jobStr := r.URL.Query().Get("job_id")

	if fromStr == "" {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "from_resume_id is required")
		return
	}
	if toStr == "" {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "to_resume_id is required")
		return
	}
	fromID, err := strconv.ParseUint(fromStr, 10, 64)
	if err != nil || fromID < 1 {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid from_resume_id")
		return
	}
	toID, err := strconv.ParseUint(toStr, 10, 64)
	if err != nil || toID < 1 {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid to_resume_id")
		return
	}
	if mode == "" { mode = "general" }
	if mode != "general" && mode != "target_direction" && mode != "target_job" {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid comparison_mode")
		return
	}
	if mode == "target_direction" && targetDir == "" {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "target_direction is required")
		return
	}
	var jobID uint64
	if mode == "target_job" {
		if jobStr == "" {
			response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "job_id is required")
			return
		}
		jobID, err = strconv.ParseUint(jobStr, 10, 64)
		if err != nil || jobID < 1 {
			response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid job_id")
			return
		}
	}

	resp, err := h.versionCompareService.GetLatestCompareV2(r.Context(), userID, fromID, toID, mode, targetDir, jobID)
	if err != nil {
		handleVersionCompareError(w, err)
		return
	}
	response.OK(w, resp)
}

func handleVersionCompareError(w http.ResponseWriter, err error) {
	if ae := apperror.AsAppError(err); ae != nil {
		response.Error(w, ae.HTTPStatus, ae.Code, ae.Message)
		return
	}
	response.Error(w, apperror.ErrInternal.HTTPStatus, apperror.ErrInternal.Code, apperror.ErrInternal.Message)
}
