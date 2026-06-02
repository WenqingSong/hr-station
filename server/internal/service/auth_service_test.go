package service

import (
	"testing"

	apperror "offer-catcher/server/internal/errors"
)

// ---------------------------------------------------------------------------
// validateEmail tests
// ---------------------------------------------------------------------------

func TestValidateEmail_Valid(t *testing.T) {
	emails := []string{
		"user@example.com",
		"test@test.co",
		"a@b.cd",
	}
	for _, e := range emails {
		if err := validateEmail(e); err != nil {
			t.Errorf("validateEmail(%q) returned error: %v", e, err)
		}
	}
}

func TestValidateEmail_Empty(t *testing.T) {
	err := validateEmail("")
	if err == nil {
		t.Fatal("validateEmail(\"\") should return error")
	}
	ae := apperror.AsAppError(err)
	if ae == nil {
		t.Fatal("validateEmail(\"\") should return AppError, not plain error")
	}
	if ae.Code != apperror.CodeInvalidParam {
		t.Errorf("code = %d, want CodeInvalidParam (%d)", ae.Code, apperror.CodeInvalidParam)
	}
}

func TestValidateEmail_InvalidFormat(t *testing.T) {
	invalids := []string{"not-an-email", "@missing-local", "missing-domain@"}
	for _, e := range invalids {
		err := validateEmail(e)
		if err == nil {
			t.Errorf("validateEmail(%q) should return error", e)
			continue
		}
		ae := apperror.AsAppError(err)
		if ae == nil {
			t.Errorf("validateEmail(%q) should return AppError, got %T", e, err)
		}
	}
}

func TestValidateEmail_SpacesOnly(t *testing.T) {
	err := validateEmail("   ")
	if err == nil {
		t.Fatal("validateEmail('   ') should return error after trim")
	}
	ae := apperror.AsAppError(err)
	if ae == nil {
		t.Fatal("should return AppError")
	}
}

// ---------------------------------------------------------------------------
// validatePassword tests
// ---------------------------------------------------------------------------

func TestValidatePassword_Valid(t *testing.T) {
	passwords := []string{"123456", "abcdef", "p@ssw0rd123"}
	for _, p := range passwords {
		if err := validatePassword(p); err != nil {
			t.Errorf("validatePassword(%q) returned error: %v", p, err)
		}
	}
}

func TestValidatePassword_Empty(t *testing.T) {
	err := validatePassword("")
	if err == nil {
		t.Fatal("validatePassword(\"\") should return error")
	}
	ae := apperror.AsAppError(err)
	if ae == nil {
		t.Fatal("should return AppError")
	}
	if ae.Code != apperror.CodeInvalidParam {
		t.Errorf("code = %d, want %d", ae.Code, apperror.CodeInvalidParam)
	}
}

func TestValidatePassword_SpacesOnly(t *testing.T) {
	err := validatePassword("   ")
	if err == nil {
		t.Fatal("validatePassword('   ') should return error after trim")
	}
	ae := apperror.AsAppError(err)
	if ae == nil {
		t.Fatal("should return AppError")
	}
}

func TestValidatePassword_TooShort(t *testing.T) {
	err := validatePassword("12345") // 5 chars, min is 6
	if err == nil {
		t.Fatal("validatePassword('12345') should return error for too-short password")
	}
	ae := apperror.AsAppError(err)
	if ae == nil {
		t.Fatal("should return AppError")
	}
	if ae.Code != apperror.CodeInvalidParam {
		t.Errorf("code = %d, want %d", ae.Code, apperror.CodeInvalidParam)
	}
}

func TestValidatePassword_ExactlyAtMinimum(t *testing.T) {
	err := validatePassword("123456") // exactly 6 chars
	if err != nil {
		t.Errorf("validatePassword('123456') should pass, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// String-matching safeguard
// ---------------------------------------------------------------------------

func TestAuthValidation_ErrorsAreAppError(t *testing.T) {
	cases := []struct {
		name string
		fn   func() error
	}{
		{"empty email", func() error { return validateEmail("") }},
		{"invalid email", func() error { return validateEmail("bad") }},
		{"empty password", func() error { return validatePassword("") }},
		{"short password", func() error { return validatePassword("12345") }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.fn()
			ae := apperror.AsAppError(err)
			if ae == nil {
				t.Fatalf("%s returned %T instead of *AppError", tc.name, err)
			}
		})
	}
}
