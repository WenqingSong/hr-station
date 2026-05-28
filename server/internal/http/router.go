package httpserver

import (
	"encoding/json"
	"net/http"
)

// NewRouter creates and returns the main HTTP handler with all routes registered.
func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthzHandler)
	return mux
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	resp := map[string]string{
		"status":  "ok",
		"service": "offer-catcher-server",
	}

	json.NewEncoder(w).Encode(resp)
}
