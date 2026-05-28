package response

import (
	"encoding/json"
	"net/http"
)

// Response represents the unified API response structure.
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// JSON writes a unified JSON response with the given HTTP status code and payload.
func JSON(w http.ResponseWriter, statusCode int, payload Response) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		// If encoding fails, attempt to write a minimal 500 response.
		http.Error(w, `{"code":50000,"message":"internal server error","data":null}`, http.StatusInternalServerError)
	}
}

// OK writes a successful response with code 0 and default message "ok".
func OK(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, Response{
		Code:    0,
		Message: "ok",
		Data:    data,
	})
}

// OKWithMessage writes a successful response with code 0 and a custom message.
func OKWithMessage(w http.ResponseWriter, message string, data interface{}) {
	JSON(w, http.StatusOK, Response{
		Code:    0,
		Message: message,
		Data:    data,
	})
}

// Error writes an error response with the given HTTP status code, business code, and message.
// Data is always null for error responses.
func Error(w http.ResponseWriter, statusCode int, code int, message string) {
	JSON(w, statusCode, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}
