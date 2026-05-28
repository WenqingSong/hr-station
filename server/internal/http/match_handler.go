package httpserver

import (
	"encoding/json"
	"net/http"

	apperror "offer-catcher/server/internal/errors"
	"offer-catcher/server/internal/middleware"
	"offer-catcher/server/internal/response"
	"offer-catcher/server/internal/service"
)

// MatchHandler handles job match analysis HTTP requests.
type MatchHandler struct {
	matchService *service.MatchService
}

// NewMatchHandler creates a new MatchHandler.
func NewMatchHandler(matchService *service.MatchService) *MatchHandler {
	return &MatchHandler{matchService: matchService}
}

// Analyze handles POST /api/jobs/{id}/match — generates mock match analysis.
func (h *MatchHandler) Analyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, apperror.ErrMethodNotAllowed.HTTPStatus, apperror.ErrMethodNotAllowed.Code, apperror.ErrMethodNotAllowed.Message)
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
		return
	}

	// jobID is passed from the router dispatcher
	jobID, ok := r.Context().Value(jobIDContextKey).(uint64)
	if !ok || jobID == 0 {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid job id")
		return
	}

	// Parse request body (optional - can be empty)
	req := service.MatchRequest{}
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

	resp, err := h.matchService.Analyze(r.Context(), userID, jobID, req)
	if err != nil {
		handleMatchError(w, err)
		return
	}

	response.OK(w, resp)
}

// handleMatchError maps match service errors to HTTP responses.
func handleMatchError(w http.ResponseWriter, err error) {
	msg := err.Error()
	switch msg {
	case "resume not found":
		response.Error(w, http.StatusNotFound, 42001, "resume not found")
	case "resume version not found":
		response.Error(w, http.StatusNotFound, 42001, "resume version not found")
	case "job not found":
		response.Error(w, http.StatusNotFound, 43001, "job not found")
	default:
		response.Error(w, apperror.ErrInternal.HTTPStatus, apperror.ErrInternal.Code, apperror.ErrInternal.Message)
	}
}
