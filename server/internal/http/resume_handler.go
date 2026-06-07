package httpserver

import (
	"context"
	"encoding/json"
	"log"
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
	resumeService             *service.ResumeService
	resumeProfileService      *service.ResumeProfileService
	optimizationV2Handler      http.HandlerFunc
	optimizationLatestHandler  http.HandlerFunc
}

// SetOptimizationV2Handler sets the handler for /api/resumes/{id}/optimization
func (h *ResumeHandler) SetOptimizationV2Handler(fn http.HandlerFunc) {
	h.optimizationV2Handler = fn
}

// SetOptimizationLatestHandler sets the handler for /api/resumes/{id}/optimization/latest
func (h *ResumeHandler) SetOptimizationLatestHandler(fn http.HandlerFunc) {
	h.optimizationLatestHandler = fn
}

// NewResumeHandler creates a new ResumeHandler.
func NewResumeHandler(resumeService *service.ResumeService, profileService *service.ResumeProfileService) *ResumeHandler {
	return &ResumeHandler{resumeService: resumeService, resumeProfileService: profileService}
}

// ── K8-5-49-B: Name check handler ──────────────────────────────────────────

func (h *ResumeHandler) CheckName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, apperror.ErrMethodNotAllowed.HTTPStatus, apperror.ErrMethodNotAllowed.Code, apperror.ErrMethodNotAllowed.Message)
		return
	}
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
		return
	}
	name := r.URL.Query().Get("name")
	excludeStr := r.URL.Query().Get("exclude_id")
	var excludeID uint64
	if excludeStr != "" {
		var err error
		excludeID, err = strconv.ParseUint(excludeStr, 10, 64)
		if err != nil {
			response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid exclude_id")
			return
		}
	}
	result, err := h.resumeService.CheckResumeName(r.Context(), userID, name, excludeID)
	if err != nil {
		handleResumeError(w, err)
		return
	}
	response.OK(w, result)
}

// ── K8-5-48-F: Relational dispatcher for /api/resumes (GET list / POST create) ──

func (h *ResumeHandler) ResumesDispatcher(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.ListRelational(w, r)
	case http.MethodPost:
		h.CreateRelational(w, r)
	default:
		response.Error(w, apperror.ErrMethodNotAllowed.HTTPStatus, apperror.ErrMethodNotAllowed.Code, apperror.ErrMethodNotAllowed.Message)
	}
}

// ── K8-5-48-F: /api/resumes/{id} sub-dispatcher ─────────────────────────────

func (h *ResumeHandler) ResumeByIDDispatcher(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/resumes/")
	if path == "" {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "resume id is required")
		return
	}

	// K9-1-H4-B: sub-routes for profile
	parts := strings.SplitN(path, "/", 3)
	idStr := parts[0]

	if len(parts) == 3 && parts[1] == "profile" && parts[2] == "status" {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id == 0 {
			response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid resume id")
			return
		}
		h.ProfileStatus(w, r)
		return
	}
	if len(parts) == 3 && parts[1] == "profile" && parts[2] == "regenerate" {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id == 0 {
			response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid resume id")
			return
		}
		h.ProfileRegenerate(w, r)
		return
	}

	// Check for /api/resumes/{id}/set-current
	if strings.HasSuffix(path, "/set-current") {
		i := strings.TrimSuffix(path, "/set-current")
		id, err := strconv.ParseUint(i, 10, 64)
		if err != nil || id == 0 {
			response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid resume id")
			return
		}
		h.SetCurrent(w, r, id)
		return
	}

	// K9-3-C: /api/resumes/{id}/optimization
	if strings.HasSuffix(path, "/optimization/latest") {
		idStr := strings.TrimSuffix(path, "/optimization/latest")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id == 0 {
			response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid resume id")
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, resumeIDContextKey, id)
		h.optimizationLatestHandler(w, r.WithContext(ctx))
		return
	}
	if strings.HasSuffix(path, "/optimization") {
		idStr := strings.TrimSuffix(path, "/optimization")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil || id == 0 {
			response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid resume id")
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, resumeIDContextKey, id)
		h.optimizationV2Handler(w, r.WithContext(ctx))
		return
	}

	id, err := strconv.ParseUint(path, 10, 64)
	if err != nil || id == 0 {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid resume id")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.DetailRelational(w, r, id)
	case http.MethodPut:
		h.UpdateRelational(w, r, id)
	case http.MethodDelete:
		h.DeleteRelational(w, r, id)
	default:
		response.Error(w, apperror.ErrMethodNotAllowed.HTTPStatus, apperror.ErrMethodNotAllowed.Code, apperror.ErrMethodNotAllowed.Message)
	}
}

// ── Relational handlers ─────────────────────────────────────────────────────

func (h *ResumeHandler) ListRelational(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
		return
	}
	resp, err := h.resumeService.List(r.Context(), userID)
	if err != nil {
		handleResumeError(w, err)
		return
	}
	response.OK(w, resp)
}

func (h *ResumeHandler) CreateRelational(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
		return
	}
	var req service.ResumeRelationalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid request body")
		return
	}
	resp, err := h.resumeService.CreateRelational(r.Context(), userID, req)
	if err != nil {
		handleResumeError(w, err)
		return
	}
	response.OK(w, resp)
}

func (h *ResumeHandler) DetailRelational(w http.ResponseWriter, r *http.Request, id uint64) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
		return
	}
	resp, err := h.resumeService.GetByIDRelational(r.Context(), userID, id)
	if err != nil {
		handleResumeError(w, err)
		return
	}
	response.OK(w, resp)
}

func (h *ResumeHandler) UpdateRelational(w http.ResponseWriter, r *http.Request, id uint64) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
		return
	}
	var req service.ResumeRelationalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid request body")
		return
	}
	resp, err := h.resumeService.UpdateRelational(r.Context(), userID, id, req)
	if err != nil {
		handleResumeError(w, err)
		return
	}
	response.OK(w, resp)
}

func (h *ResumeHandler) SetCurrent(w http.ResponseWriter, r *http.Request, id uint64) {
	if r.Method != http.MethodPost {
		response.Error(w, apperror.ErrMethodNotAllowed.HTTPStatus, apperror.ErrMethodNotAllowed.Code, apperror.ErrMethodNotAllowed.Message)
		return
	}
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
		return
	}
	resp, err := h.resumeService.SetCurrentRelational(r.Context(), userID, id)
	if err != nil {
		handleResumeError(w, err)
		return
	}
	response.OK(w, resp)
}

func (h *ResumeHandler) DeleteRelational(w http.ResponseWriter, r *http.Request, id uint64) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
		return
	}
	_, err := h.resumeService.Delete(r.Context(), userID, id)
	if err != nil {
		handleResumeError(w, err)
		return
	}
	response.OK(w, map[string]string{"status": "deleted"})
}

// ── K9-1-H4-B: Profile status & regenerate handlers ─────────────────────────

func (h *ResumeHandler) ProfileStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, apperror.ErrMethodNotAllowed.HTTPStatus, apperror.ErrMethodNotAllowed.Code, apperror.ErrMethodNotAllowed.Message)
		return
	}
	resumeID := extractProfileResumeID(r)
	if resumeID == 0 {
		return
	}
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
		return
	}
	resp, err := h.resumeProfileService.GetProfileStatus(r.Context(), userID, resumeID)
	if err != nil {
		handleResumeError(w, err)
		return
	}
	response.OK(w, resp)
}

func (h *ResumeHandler) ProfileRegenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, apperror.ErrMethodNotAllowed.HTTPStatus, apperror.ErrMethodNotAllowed.Code, apperror.ErrMethodNotAllowed.Message)
		return
	}
	resumeID := extractProfileResumeID(r)
	if resumeID == 0 {
		return
	}
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
		return
	}
	var req struct {
		ProfileType string `json:"profile_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid request body")
		return
	}
	if err := h.resumeProfileService.RegenerateProfile(r.Context(), userID, resumeID, req.ProfileType); err != nil {
		log.Printf("[ResumeHandler] ProfileRegenerate resume=%d err=%v", resumeID, err)
		handleResumeError(w, err)
		return
	}
	response.OK(w, map[string]interface{}{
		"resume_id":    resumeID,
		"profile_type": req.ProfileType,
		"status":       "pending",
	})
}

func extractProfileResumeID(r *http.Request) uint64 {
	path := strings.TrimPrefix(r.URL.Path, "/api/resumes/")
	parts := strings.SplitN(path, "/", 2)
	id, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil || id == 0 {
		return 0
	}
	return id
}

// ── @deprecated: old V1/V2 handlers (kept for backward compat) ──────────────

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

// handleResumeError maps resume service errors to HTTP responses using errors.As.
func handleResumeError(w http.ResponseWriter, err error) {
	if ae := apperror.AsAppError(err); ae != nil {
		response.Error(w, ae.HTTPStatus, ae.Code, ae.Message)
		return
	}
	response.Error(w, apperror.ErrInternal.HTTPStatus, apperror.ErrInternal.Code, apperror.ErrInternal.Message)
}
