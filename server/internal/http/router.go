package httpserver

import (
	"net/http"

	"offer-catcher/server/internal/errors"
	"offer-catcher/server/internal/response"
)

// NewRouter creates and returns the main HTTP handler with all routes registered.
func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthzHandler)
	return mux
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, apperror.ErrMethodNotAllowed.HTTPStatus, apperror.ErrMethodNotAllowed.Code, apperror.ErrMethodNotAllowed.Message)
		return
	}

	data := map[string]string{
		"status":  "ok",
		"service": "offer-catcher-server",
	}

	response.OK(w, data)
}
