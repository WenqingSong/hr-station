package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError represents an application-level error with code, message, and HTTP status.
type AppError struct {
	Code       int
	Message    string
	HTTPStatus int
	Err        error
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("code=%d message=%s err=%v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("code=%d message=%s", e.Code, e.Message)
}

// Unwrap returns the underlying error for use with errors.Is/As.
func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError with the given code, message, and HTTP status.
func New(code int, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// Wrap creates a new AppError wrapping an existing error.
func Wrap(code int, message string, httpStatus int, err error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Err:        err,
	}
}

// AsAppError extracts an *AppError from err via errors.As.
func AsAppError(err error) *AppError {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae
	}
	return nil
}

// Base error codes
const (
	CodeOK                    = 0
	CodeBadRequest            = 40000
	CodeInvalidParam          = 40001
	CodeUnauthorized          = 40100
	CodeForbidden             = 40300
	CodeNotFound              = 40400
	CodeMethodNotAllowed      = 40500
	CodeInternal              = 50000
	CodeDependencyUnavailable = 50300
)

// Auth module codes (410xx)
const (
	CodeEmailRegistered  = 41001
	CodeInvalidCreds     = 41002
)

// Resume module codes (420xx)
const (
	CodeResumeNotFound        = 42001
	CodeResumeNotAvailable    = 42002
	CodeResumeFromVersionNotFound = 42003
	CodeResumeToVersionNotFound   = 42004
)

// Jobs / Agent module codes (430xx)
const (
	CodeJobNotFound = 43001
)

// Predefined base errors
var (
	ErrBadRequest            = New(CodeBadRequest, "bad request", http.StatusBadRequest)
	ErrInvalidParam          = New(CodeInvalidParam, "invalid parameter", http.StatusBadRequest)
	ErrUnauthorized          = New(CodeUnauthorized, "unauthorized", http.StatusUnauthorized)
	ErrForbidden             = New(CodeForbidden, "forbidden", http.StatusForbidden)
	ErrNotFound              = New(CodeNotFound, "not found", http.StatusNotFound)
	ErrMethodNotAllowed      = New(CodeMethodNotAllowed, "method not allowed", http.StatusMethodNotAllowed)
	ErrInternal              = New(CodeInternal, "internal server error", http.StatusInternalServerError)
	ErrDependencyUnavailable = New(CodeDependencyUnavailable, "dependency unavailable", http.StatusServiceUnavailable)
)

// Auth errors
var (
	ErrEmailRegistered = New(CodeEmailRegistered, "email already registered", http.StatusBadRequest)
	ErrInvalidCreds    = New(CodeInvalidCreds, "invalid email or password", http.StatusUnauthorized)
)

// Resume errors
var (
	ErrResumeNotFound          = New(CodeResumeNotFound, "resume version not found", http.StatusNotFound)
	ErrResumeNotAvailable      = New(CodeResumeNotAvailable, "resume not found", http.StatusNotFound)
	ErrResumeFromVerNotFound   = New(CodeResumeFromVersionNotFound, "from version not found", http.StatusNotFound)
	ErrResumeToVerNotFound     = New(CodeResumeToVersionNotFound, "to version not found", http.StatusNotFound)
)

// Job errors
var (
	ErrJobNotFound = New(CodeJobNotFound, "job not found", http.StatusNotFound)
)
