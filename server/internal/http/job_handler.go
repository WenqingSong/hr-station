package httpserver

import (
	"net/http"
	"strconv"
	"strings"

	apperror "offer-catcher/server/internal/errors"
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

// List handles GET /api/jobs — returns a paginated list of jobs (public).
func (h *JobHandler) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, apperror.ErrMethodNotAllowed.HTTPStatus, apperror.ErrMethodNotAllowed.Code, apperror.ErrMethodNotAllowed.Message)
		return
	}

	q := r.URL.Query()
	keyword := q.Get("keyword")
	direction := q.Get("direction")
	city := q.Get("city")
	recruitType := q.Get("recruit_type")
	category := q.Get("category")
	externalSource := q.Get("external_source")

	// Support both page/page_size (frontend) and limit/offset (backend)
	page := 1
	pageSize := 10
	if v := q.Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	if v := q.Get("page_size"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			pageSize = n
		}
	}

	// Backward-compatible: limit/offset override page/page_size
	limit := pageSize
	offset := (page - 1) * pageSize
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	resp, err := h.jobService.List(r.Context(), service.JobListRequest{
		Keyword:        keyword,
		Direction:      direction,
		City:           city,
		RecruitType:    recruitType,
		Category:       category,
		ExternalSource: externalSource,
		Limit:          limit,
		Offset:         offset,
	})
	if err != nil {
		handleJobError(w, err)
		return
	}

	response.OK(w, resp)
}

// Detail handles GET /api/jobs/{id} — returns a single job detail (public).
func (h *JobHandler) Detail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, apperror.ErrMethodNotAllowed.HTTPStatus, apperror.ErrMethodNotAllowed.Code, apperror.ErrMethodNotAllowed.Message)
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

// Facets handles GET /api/jobs/facets — returns aggregated facet counts (public).
func (h *JobHandler) Facets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, apperror.ErrMethodNotAllowed.HTTPStatus, apperror.ErrMethodNotAllowed.Code, apperror.ErrMethodNotAllowed.Message)
		return
	}

	resp, err := h.jobService.GetFacets(r.Context())
	if err != nil {
		response.Error(w, apperror.ErrInternal.HTTPStatus, apperror.ErrInternal.Code, apperror.ErrInternal.Message)
		return
	}

	response.OK(w, resp)
}

// handleJobError maps job service errors to HTTP responses using errors.As.
func handleJobError(w http.ResponseWriter, err error) {
	if ae := apperror.AsAppError(err); ae != nil {
		response.Error(w, ae.HTTPStatus, ae.Code, ae.Message)
		return
	}
	response.Error(w, apperror.ErrInternal.HTTPStatus, apperror.ErrInternal.Code, apperror.ErrInternal.Message)
}
