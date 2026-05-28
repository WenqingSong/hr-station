package httpserver

import (
	"net/http"
	"strconv"
	"strings"

	apperror "offer-catcher/server/internal/errors"
	"offer-catcher/server/internal/middleware"
	"offer-catcher/server/internal/response"
	"offer-catcher/server/internal/service"
)

// JobHandler handles job HTTP requests.
type JobHandler struct {
	jobService *service.JobService
}

// NewJobHandler creates a new JobHandler.
func NewJobHandler(jobService *service.JobService) *JobHandler {
	return &JobHandler{jobService: jobService}
}

// List handles GET /api/jobs — returns a paginated list of jobs.
func (h *JobHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, apperror.ErrMethodNotAllowed.HTTPStatus, apperror.ErrMethodNotAllowed.Code, apperror.ErrMethodNotAllowed.Message)
		return
	}

	if _, ok := middleware.UserIDFromContext(r.Context()); !ok {
		response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
		return
	}

	keyword := r.URL.Query().Get("keyword")
	city := r.URL.Query().Get("city")
	category := r.URL.Query().Get("category")

	limit := 20
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}

	offset := 0
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	resp, err := h.jobService.List(r.Context(), service.JobListRequest{
		Keyword:  keyword,
		City:     city,
		Category: category,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		response.Error(w, apperror.ErrInternal.HTTPStatus, apperror.ErrInternal.Code, apperror.ErrInternal.Message)
		return
	}

	response.OK(w, resp)
}

// Detail handles GET /api/jobs/{id} — returns a single job detail.
func (h *JobHandler) Detail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, apperror.ErrMethodNotAllowed.HTTPStatus, apperror.ErrMethodNotAllowed.Code, apperror.ErrMethodNotAllowed.Message)
		return
	}

	if _, ok := middleware.UserIDFromContext(r.Context()); !ok {
		response.Error(w, apperror.ErrUnauthorized.HTTPStatus, apperror.ErrUnauthorized.Code, apperror.ErrUnauthorized.Message)
		return
	}

	// Extract job ID from path: /api/jobs/{id}
	path := strings.TrimPrefix(r.URL.Path, "/api/jobs/")
	if path == "" {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "job id is required")
		return
	}

	id, err := strconv.ParseUint(path, 10, 64)
	if err != nil || id == 0 {
		response.Error(w, apperror.ErrInvalidParam.HTTPStatus, apperror.ErrInvalidParam.Code, "invalid job id")
		return
	}

	resp, err := h.jobService.GetDetail(r.Context(), id)
	if err != nil {
		handleJobError(w, err)
		return
	}

	response.OK(w, resp)
}

// handleJobError maps job service errors to HTTP responses.
func handleJobError(w http.ResponseWriter, err error) {
	msg := err.Error()
	switch msg {
	case "job not found":
		response.Error(w, http.StatusNotFound, 43001, "job not found")
	default:
		response.Error(w, apperror.ErrInternal.HTTPStatus, apperror.ErrInternal.Code, apperror.ErrInternal.Message)
	}
}
