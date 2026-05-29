package httpserver

import (
	"net/http"
	"strconv"

	apperror "offer-catcher/server/internal/errors"
	"offer-catcher/server/internal/middleware"
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

	// Parse query parameters
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if fromStr == "" {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "from parameter is required")
		return
	}
	if toStr == "" {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "to parameter is required")
		return
	}

	fromVersionNo, err := strconv.Atoi(fromStr)
	if err != nil || fromVersionNo < 1 {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid from version number")
		return
	}

	toVersionNo, err := strconv.Atoi(toStr)
	if err != nil || toVersionNo < 1 {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid to version number")
		return
	}

	if fromVersionNo == toVersionNo {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "from and to version numbers must be different")
		return
	}

	req := service.VersionCompareRequest{
		FromVersionNo: fromVersionNo,
		ToVersionNo:   toVersionNo,
	}

	resp, err := h.versionCompareService.Compare(r.Context(), userID, req)
	if err != nil {
		handleVersionCompareError(w, err)
		return
	}

	response.OK(w, resp)
}

// handleVersionCompareError maps version compare service errors to HTTP responses.
func handleVersionCompareError(w http.ResponseWriter, err error) {
	msg := err.Error()
	switch msg {
	case "invalid from version number":
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid from version number")
	case "invalid to version number":
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid to version number")
	case "from and to version numbers must be different":
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "from and to version numbers must be different")
	case "from version not found":
		response.Error(w, http.StatusNotFound, 42001, "resume version not found")
	case "to version not found":
		response.Error(w, http.StatusNotFound, 42001, "resume version not found")
	default:
		response.Error(w, apperror.ErrInternal.HTTPStatus, apperror.ErrInternal.Code, apperror.ErrInternal.Message)
	}
}
