package httpserver

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	apperror "offer-catcher/server/internal/errors"
	"offer-catcher/server/internal/middleware"
	"offer-catcher/server/internal/response"
	"offer-catcher/server/internal/service"
)

// ResumeHandler handles resume HTTP requests.
type ResumeHandler struct {
	resumeService *service.ResumeService
}

// NewResumeHandler creates a new ResumeHandler.
func NewResumeHandler(resumeService *service.ResumeService) *ResumeHandler {
	return &ResumeHandler{resumeService: resumeService}
}

// Create handles POST /api/resumes — creates a new resume version.
func (h *ResumeHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, apperror.ErrMethodNotAllowed.HTTPStatus, apperror.ErrMethodNotAllowed.Code, apperror.ErrMethodNotAllowed.Message)
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
		return
	}

	var req service.CreateResumeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid request body")
		return
	}

	resp, err := h.resumeService.CreateVersion(r.Context(), userID, req)
	if err != nil {
		handleResumeError(w, err)
		return
	}

	response.OK(w, resp)
}

// Current handles GET /api/resumes/current — returns the latest resume version.
func (h *ResumeHandler) Current(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, apperror.ErrMethodNotAllowed.HTTPStatus, apperror.ErrMethodNotAllowed.Code, apperror.ErrMethodNotAllowed.Message)
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
		return
	}

	resp, err := h.resumeService.GetCurrent(r.Context(), userID)
	if err != nil {
		handleResumeError(w, err)
		return
	}

	response.OK(w, resp)
}

// Detail handles GET /api/resumes/detail — returns the latest resume version (same as current).
func (h *ResumeHandler) Detail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, apperror.ErrMethodNotAllowed.HTTPStatus, apperror.ErrMethodNotAllowed.Code, apperror.ErrMethodNotAllowed.Message)
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
		return
	}

	resp, err := h.resumeService.GetCurrent(r.Context(), userID)
	if err != nil {
		handleResumeError(w, err)
		return
	}

	response.OK(w, resp)
}

// Versions handles GET /api/resumes/versions — returns all resume versions.
func (h *ResumeHandler) Versions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, apperror.ErrMethodNotAllowed.HTTPStatus, apperror.ErrMethodNotAllowed.Code, apperror.ErrMethodNotAllowed.Message)
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
		return
	}

	resp, err := h.resumeService.ListVersions(r.Context(), userID)
	if err != nil {
		handleResumeError(w, err)
		return
	}

	response.OK(w, resp)
}

// VersionDetail handles GET /api/resumes/versions/{versionNo} — returns a specific version.
// The version number is parsed from the URL path after "/api/resumes/versions/".
func (h *ResumeHandler) VersionDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, apperror.ErrMethodNotAllowed.HTTPStatus, apperror.ErrMethodNotAllowed.Code, apperror.ErrMethodNotAllowed.Message)
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
		return
	}

	// Extract version number from path: /api/resumes/versions/{versionNo}
	path := strings.TrimPrefix(r.URL.Path, "/api/resumes/versions/")
	if path == "" {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "version number is required")
		return
	}

	versionNo, err := strconv.Atoi(path)
	if err != nil || versionNo < 1 {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid version number")
		return
	}

	resp, err := h.resumeService.GetByVersionNo(r.Context(), userID, versionNo)
	if err != nil {
		handleResumeError(w, err)
		return
	}

	response.OK(w, resp)
}

// handleResumeError maps resume service errors to HTTP responses.
func handleResumeError(w http.ResponseWriter, err error) {
	msg := err.Error()
	switch msg {
	case "invalid version number":
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, msg)
	case "resume version not found":
		response.Error(w, http.StatusNotFound, 42001, "resume version not found")
	default:
		response.Error(w, apperror.ErrInternal.HTTPStatus, apperror.ErrInternal.Code, apperror.ErrInternal.Message)
	}
}
