package apperror

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

// ---------------------------------------------------------------------------
// AppError Creation & Properties
// ---------------------------------------------------------------------------

func TestNew_AppError_Basic(t *testing.T) {
	ae := New(40001, "invalid parameter", http.StatusBadRequest)
	if ae.Code != 40001 {
		t.Errorf("New().Code = %d, want 40001", ae.Code)
	}
	if ae.Message != "invalid parameter" {
		t.Errorf("New().Message = %q, want %q", ae.Message, "invalid parameter")
	}
	if ae.HTTPStatus != http.StatusBadRequest {
		t.Errorf("New().HTTPStatus = %d, want %d", ae.HTTPStatus, http.StatusBadRequest)
	}
	if ae.Err != nil {
		t.Error("New().Err should be nil for basic creation")
	}
}

func TestNew_AppError_ZeroValues(t *testing.T) {
	ae := New(0, "", 0)
	if ae.Code != 0 {
		t.Errorf("zero code should be 0, got %d", ae.Code)
	}
	if ae.Message != "" {
		t.Errorf("zero message should be empty, got %q", ae.Message)
	}
}

func TestWrap_AppError_WithUnderlying(t *testing.T) {
	underlying := fmt.Errorf("database timeout")
	ae := Wrap(50000, "internal error", http.StatusInternalServerError, underlying)
	if ae.Err == nil {
		t.Fatal("Wrap().Err should not be nil")
	}
	if ae.Err != underlying {
		t.Errorf("Wrap().Err should be the underlying error")
	}
}

// ---------------------------------------------------------------------------
// Error() interface
// ---------------------------------------------------------------------------

func TestAppError_Error_WithErr(t *testing.T) {
	ae := Wrap(50000, "internal", http.StatusInternalServerError, fmt.Errorf("db timeout"))
	s := ae.Error()
	expected := "code=50000 message=internal err=db timeout"
	if s != expected {
		t.Errorf("Error() = %q, want %q", s, expected)
	}
}

func TestAppError_Error_WithoutErr(t *testing.T) {
	ae := New(40001, "invalid param", http.StatusBadRequest)
	s := ae.Error()
	expected := "code=40001 message=invalid param"
	if s != expected {
		t.Errorf("Error() = %q, want %q", s, expected)
	}
}

// ---------------------------------------------------------------------------
// errors.As / AsAppError
// ---------------------------------------------------------------------------

func TestAsAppError_RecognizesAppError(t *testing.T) {
	ae := New(40001, "invalid", 400)
	extracted := AsAppError(ae)
	if extracted == nil {
		t.Fatal("AsAppError returned nil for bare AppError")
	}
	if extracted.Code != 40001 {
		t.Errorf("AsAppError lost code: got %d", extracted.Code)
	}
}

func TestAsAppError_RecognizesWrappedAppError(t *testing.T) {
	ae := New(40001, "invalid", 400)
	wrapped := fmt.Errorf("wrapped: %w", ae)
	extracted := AsAppError(wrapped)
	if extracted == nil {
		t.Fatal("AsAppError returned nil for wrapped AppError")
	}
	if extracted.Message != "invalid" {
		t.Errorf("AsAppError lost message: got %q", extracted.Message)
	}
}

func TestAsAppError_NonAppError(t *testing.T) {
	plain := errors.New("something went wrong")
	extracted := AsAppError(plain)
	if extracted != nil {
		t.Error("AsAppError should return nil for non-AppError")
	}
}

func TestAsAppError_NilError(t *testing.T) {
	extracted := AsAppError(nil)
	if extracted != nil {
		t.Error("AsAppError should return nil for nil error")
	}
}

// ---------------------------------------------------------------------------
// Unwrap
// ---------------------------------------------------------------------------

func TestAppError_Unwrap(t *testing.T) {
	underlying := fmt.Errorf("root cause")
	ae := Wrap(50000, "internal", 500, underlying)
	unwrapped := errors.Unwrap(ae)
	if unwrapped != underlying {
		t.Errorf("Unwrap should return underlying error")
	}
}

func TestAppError_ErrorsIs(t *testing.T) {
	sentinel := errors.New("sentinel")
	ae := Wrap(50000, "internal", 500, sentinel)
	if !errors.Is(ae, sentinel) {
		t.Error("errors.Is should find sentinel through AppError.Wrap")
	}
}

// ---------------------------------------------------------------------------
// Sentinel errors — existence and properties
// ---------------------------------------------------------------------------

func TestSentinelErrors_Exist(t *testing.T) {
	tests := []struct {
		name     string
		err      *AppError
		wantCode int
		wantHTTP int
	}{
		{"ErrBadRequest", ErrBadRequest, 40000, 400},
		{"ErrInvalidParam", ErrInvalidParam, 40001, 400},
		{"ErrUnauthorized", ErrUnauthorized, 40100, 401},
		{"ErrForbidden", ErrForbidden, 40300, 403},
		{"ErrNotFound", ErrNotFound, 40400, 404},
		{"ErrMethodNotAllowed", ErrMethodNotAllowed, 40500, 405},
		{"ErrInternal", ErrInternal, 50000, 500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Fatal("sentinel is nil")
			}
			if tt.err.Code != tt.wantCode {
				t.Errorf("%s.Code = %d, want %d", tt.name, tt.err.Code, tt.wantCode)
			}
			if tt.err.HTTPStatus != tt.wantHTTP {
				t.Errorf("%s.HTTPStatus = %d, want %d", tt.name, tt.err.HTTPStatus, tt.wantHTTP)
			}
		})
	}
}

func TestAuthSentinelErrors(t *testing.T) {
	if ErrEmailRegistered.Code != 41001 {
		t.Errorf("ErrEmailRegistered.Code = %d, want 41001", ErrEmailRegistered.Code)
	}
	if ErrInvalidCreds.Code != 41002 {
		t.Errorf("ErrInvalidCreds.Code = %d, want 41002", ErrInvalidCreds.Code)
	}
}

func TestResumeSentinelErrors(t *testing.T) {
	if ErrResumeNotFound.Code != 42001 {
		t.Errorf("ErrResumeNotFound.Code = %d, want 42001", ErrResumeNotFound.Code)
	}
	if ErrResumeNotAvailable.Code != 42002 {
		t.Errorf("ErrResumeNotAvailable.Code = %d, want 42002", ErrResumeNotAvailable.Code)
	}
	if ErrResumeFromVerNotFound.Code != 42003 {
		t.Errorf("ErrResumeFromVerNotFound.Code = %d, want 42003", ErrResumeFromVerNotFound.Code)
	}
	if ErrResumeToVerNotFound.Code != 42004 {
		t.Errorf("ErrResumeToVerNotFound.Code = %d, want 42004", ErrResumeToVerNotFound.Code)
	}
}

func TestJobSentinelErrors(t *testing.T) {
	if ErrJobNotFound.Code != 43001 {
		t.Errorf("ErrJobNotFound.Code = %d, want 43001", ErrJobNotFound.Code)
	}
	if ErrInvalidExternalSource == nil {
		t.Fatal("ErrInvalidExternalSource is nil")
	}
	if ErrInvalidExternalSource.Code != CodeInvalidParam {
		t.Errorf("ErrInvalidExternalSource.Code = %d, want CodeInvalidParam (%d)", ErrInvalidExternalSource.Code, CodeInvalidParam)
	}
	if ErrInvalidExternalSource.Message != "invalid external_source" {
		t.Errorf("ErrInvalidExternalSource.Message = %q", ErrInvalidExternalSource.Message)
	}
	if ErrInvalidExternalSource.HTTPStatus != http.StatusBadRequest {
		t.Errorf("ErrInvalidExternalSource.HTTPStatus = %d, want 400", ErrInvalidExternalSource.HTTPStatus)
	}
}

// ---------------------------------------------------------------------------
// AsAppError on specific sentinel wrappers
// ---------------------------------------------------------------------------

func TestAsAppError_SentinelWrapped(t *testing.T) {
	// Simulate service returning fmt.Errorf("...", %w, ErrInvalidExternalSource)
	orig := ErrInvalidExternalSource
	wrapped := fmt.Errorf("validation failed: %w", orig)
	extracted := AsAppError(wrapped)
	if extracted == nil {
		t.Fatal("AsAppError did not find ErrInvalidExternalSource in wrapped error")
	}
	if extracted.Code != 40001 {
		t.Errorf("extracted code = %d, want 40001", extracted.Code)
	}
}

func TestAsAppError_DeepWrapping(t *testing.T) {
	// Simulate: repository error → service wraps → handler receives
	ae := ErrInternal
	w1 := fmt.Errorf("service: %w", ae)
	w2 := fmt.Errorf("handler: %w", w1)
	extracted := AsAppError(w2)
	if extracted == nil {
		t.Fatal("AsAppError did not find AppError through deep wrapping")
	}
	if extracted.Code != 50000 {
		t.Errorf("extracted code = %d, want 50000", extracted.Code)
	}
}

// ---------------------------------------------------------------------------
// Error code namespace — no overlaps
// ---------------------------------------------------------------------------

func TestErrorCodes_NoOverlap(t *testing.T) {
	seen := map[int]string{}
	tests := []struct {
		name string
		code int
	}{
		{"CodeBadRequest", CodeBadRequest},
		{"CodeInvalidParam", CodeInvalidParam},
		{"CodeUnauthorized", CodeUnauthorized},
		{"CodeForbidden", CodeForbidden},
		{"CodeNotFound", CodeNotFound},
		{"CodeMethodNotAllowed", CodeMethodNotAllowed},
		{"CodeInternal", CodeInternal},
		{"CodeEmailRegistered", CodeEmailRegistered},
		{"CodeInvalidCreds", CodeInvalidCreds},
		{"CodeResumeNotFound", CodeResumeNotFound},
		{"CodeResumeNotAvailable", CodeResumeNotAvailable},
		{"CodeResumeFromVersionNotFound", CodeResumeFromVersionNotFound},
		{"CodeResumeToVersionNotFound", CodeResumeToVersionNotFound},
		{"CodeJobNotFound", CodeJobNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if prev, exists := seen[tt.code]; exists {
				t.Errorf("code %d overlaps: %s and %s share the same code", tt.code, prev, tt.name)
			}
			seen[tt.code] = tt.name
		})
	}
}
