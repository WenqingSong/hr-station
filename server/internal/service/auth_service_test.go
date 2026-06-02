package service

import (
	"testing"

	apperror "offer-catcher/server/internal/errors"
)

// ---------------------------------------------------------------------------
// validatePhone tests
// ---------------------------------------------------------------------------

func TestValidatePhone_Valid(t *testing.T) {
	phones := []string{
		"13800138000",
		"13900001001",
		"18812345678",
	}
	for _, p := range phones {
		if err := validatePhone(p); err != nil {
			t.Errorf("validatePhone(%q) returned error: %v", p, err)
		}
	}
}

func TestValidatePhone_Empty(t *testing.T) {
	err := validatePhone("")
	if err == nil {
		t.Fatal("validatePhone(\"\") should return error")
	}
	ae := apperror.AsAppError(err)
	if ae == nil {
		t.Fatal("validatePhone(\"\") should return AppError, not plain error")
	}
	if ae.Code != apperror.CodeInvalidParam {
		t.Errorf("code = %d, want CodeInvalidParam (%d)", ae.Code, apperror.CodeInvalidParam)
	}
}

func TestValidatePhone_InvalidFormat(t *testing.T) {
	invalids := []string{"12345", "12345678901", "22345678901", "abc1234567"}
	for _, p := range invalids {
		err := validatePhone(p)
		if err == nil {
			t.Errorf("validatePhone(%q) should return error", p)
			continue
		}
		ae := apperror.AsAppError(err)
		if ae == nil {
			t.Errorf("validatePhone(%q) should return AppError, got %T", p, err)
		}
	}
}

func TestValidatePhone_SpacesOnly(t *testing.T) {
	err := validatePhone("   ")
	if err == nil {
		t.Fatal("validatePhone('   ') should return error after trim")
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
		{"empty phone", func() error { return validatePhone("") }},
		{"invalid phone", func() error { return validatePhone("abc") }},
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
