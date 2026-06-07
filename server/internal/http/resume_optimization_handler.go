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

type ResumeOptimizationHandler struct {
	resumeOptimizationService *service.ResumeOptimizationService
}

func NewResumeOptimizationHandler(s *service.ResumeOptimizationService) *ResumeOptimizationHandler {
	return &ResumeOptimizationHandler{resumeOptimizationService: s}
}

func (h *ResumeOptimizationHandler) Optimize(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getLatest(w, r)
	case http.MethodPost:
		h.generate(w, r)
	default:
		response.Error(w, apperror.ErrMethodNotAllowed.HTTPStatus, apperror.ErrMethodNotAllowed.Code, apperror.ErrMethodNotAllowed.Message)
	}
}

func (h *ResumeOptimizationHandler) OptimizeV2(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, apperror.ErrMethodNotAllowed.HTTPStatus, apperror.ErrMethodNotAllowed.Code, apperror.ErrMethodNotAllowed.Message)
		return
	}
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
		return
	}
	resumeID, ok := r.Context().Value(resumeIDContextKey).(uint64)
	if !ok || resumeID == 0 {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid resume id")
		return
	}
	var req service.OptimizationV2Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid request body")
		return
	}
	resp, err := h.resumeOptimizationService.OptimizeV2(r.Context(), userID, resumeID, req)
	if err != nil {
		handleResumeOptimizationError(w, err)
		return
	}
	response.OK(w, resp)
}

// LatestV2 handles GET /api/resumes/{id}/optimization/latest (K9-3-D)
func (h *ResumeOptimizationHandler) LatestV2(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, apperror.ErrMethodNotAllowed.HTTPStatus, apperror.ErrMethodNotAllowed.Code, apperror.ErrMethodNotAllowed.Message)
		return
	}
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
		return
	}
	resumeID, ok := r.Context().Value(resumeIDContextKey).(uint64)
	if !ok || resumeID == 0 {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid resume id")
		return
	}
	q := r.URL.Query()
	mode := q.Get("optimization_mode")
	if mode == "" { mode = "general" }
	targetDir := q.Get("target_direction")
	jobIDStr := q.Get("job_id")
	var jobID uint64
	if jobIDStr != "" {
		var err error
		jobID, err = strconv.ParseUint(jobIDStr, 10, 64)
		if err != nil {
			response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid job_id")
			return
		}
	}
	resp, err := h.resumeOptimizationService.GetLatestOptimizationV2(r.Context(), userID, resumeID, mode, targetDir, jobID)
	if err != nil {
		handleResumeOptimizationError(w, err)
		return
	}
	response.OK(w, resp)
}

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

func (h *ResumeOptimizationHandler) generate(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
		return
	}
	var req service.ResumeOptimizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid request body")
		return
	}
	tier, err := model.NormalizeGenerationTier(string(req.GenerationTier))
	if err != nil {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, err.Error())
		return
	}
	req.GenerationTier = tier
	resp, err := h.resumeOptimizationService.Optimize(r.Context(), userID, req)
	if err != nil {
		handleResumeOptimizationError(w, err)
		return
	}
	response.OK(w, resp)
}

func handleResumeOptimizationError(w http.ResponseWriter, err error) {
	if ae := apperror.AsAppError(err); ae != nil {
		response.Error(w, ae.HTTPStatus, ae.Code, ae.Message)
		return
	}
	response.Error(w, apperror.ErrInternal.HTTPStatus, apperror.ErrInternal.Code, apperror.ErrInternal.Message)
}
