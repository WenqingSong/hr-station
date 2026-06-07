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

func handleVersionCompareError(w http.ResponseWriter, err error) {
	if ae := apperror.AsAppError(err); ae != nil {
		response.Error(w, ae.HTTPStatus, ae.Code, ae.Message)
		return
	}
	response.Error(w, apperror.ErrInternal.HTTPStatus, apperror.ErrInternal.Code, apperror.ErrInternal.Message)
}
