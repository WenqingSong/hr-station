package llm

import "testing"

// TestHy3ErrorFromHTTPStatus verifies that HTTP status codes are mapped to
// accurate error codes. In particular, 401/403 must map to ErrLLMAuthFailed
// (auth failure: key invalid / expired / platform mismatch) rather than
// ErrLLMAPIKeyMissing (empty/missing key), so logs and alerts don't mislead
// operators into thinking the key was simply not configured.
func TestHy3ErrorFromHTTPStatus(t *testing.T) {
	tests := []struct {
		code int
		want ErrorCode
	}{
		{401, ErrLLMAuthFailed},
		{403, ErrLLMAuthFailed},
		{429, ErrLLMRateLimit},
		{500, ErrLLMHTTPError},
		{502, ErrLLMHTTPError},
		{503, ErrLLMHTTPError},
		{400, ErrLLMHTTPError},
		{404, ErrLLMHTTPError},
	}
	for _, tt := range tests {
		got := errorFromHTTPStatus(tt.code)
		if got != tt.want {
			t.Errorf("errorFromHTTPStatus(%d) = %q, want %q", tt.code, got, tt.want)
		}
	}
}
