package response

import (
	"encoding/json"
	"net/http"

	apperror "offer-catcher/server/internal/errors"
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
func Error(w http.ResponseWriter, statusCode int, code int, message string) {
	JSON(w, statusCode, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// WriteAppError extracts an *apperror.AppError from err and writes the appropriate
// error response. If err is not an AppError, it falls back to ErrInternal.
func WriteAppError(w http.ResponseWriter, err error) {
	if ae := apperror.AsAppError(err); ae != nil {
		Error(w, ae.HTTPStatus, ae.Code, ae.Message)
		return
	}
	Error(w, apperror.ErrInternal.HTTPStatus, apperror.ErrInternal.Code, apperror.ErrInternal.Message)
}
