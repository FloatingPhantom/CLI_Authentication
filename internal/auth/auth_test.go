package auth

import (
	"testing"
)

func TestErrTOTPRequired(t *testing.T) {
	err := &ErrTOTPRequired{
		UserID:   42,
		Username: "testuser",
	}

	// Should implement the error interface
	var _ error = err

	if err.Error() != "TOTP code required" {
		t.Errorf("ErrTOTPRequired.Error() = %q, want %q", err.Error(), "TOTP code required")
	}
	if err.UserID != 42 {
		t.Errorf("ErrTOTPRequired.UserID = %d, want 42", err.UserID)
	}
	if err.Username != "testuser" {
		t.Errorf("ErrTOTPRequired.Username = %q, want %q", err.Username, "testuser")
	}
}

func TestTOTPEnrollment(t *testing.T) {
	enrollment, err := NewService(nil, nil, nil).EnableTOTP("testuser")
	if err != nil {
		t.Fatalf("EnableTOTP() unexpected error: %v", err)
	}

	if enrollment.Secret == "" {
		t.Error("TOTPEnrollment.Secret is empty")
	}
	if enrollment.URL == "" {
		t.Error("TOTPEnrollment.URL is empty")
	}
	if enrollment.Key == nil {
		t.Error("TOTPEnrollment.Key is nil")
	}
}

func TestLoginResult_Struct(t *testing.T) {
	result := &LoginResult{}
	if result.User != nil {
		t.Error("LoginResult.User should be nil by default")
	}
	if result.Token != "" {
		t.Error("LoginResult.Token should be empty by default")
	}
	if result.ExpiresAt.IsZero() != true {
		t.Error("LoginResult.ExpiresAt should be zero by default")
	}
}

func TestNewService(t *testing.T) {
	svc := NewService(nil, nil, nil)
	if svc == nil {
		t.Fatal("NewService() returned nil")
	}
}

func TestRegister_Validation(t *testing.T) {
	svc := NewService(nil, nil, nil)

	tests := []struct {
		name     string
		username string
		password string
		wantErr  string
	}{
		{
			name:     "empty username",
			username: "",
			password: "Str0ng!Pass",
			wantErr:  "username cannot be empty",
		},
		{
			name:     "whitespace only username",
			username: "   ",
			password: "Str0ng!Pass",
			wantErr:  "username cannot be empty",
		},
		{
			name:     "username too short",
			username: "ab",
			password: "Str0ng!Pass",
			wantErr:  "username must be at least 3 characters",
		},
		{
			name:     "username too long",
			username: string(make([]byte, 65)),
			password: "Str0ng!Pass",
			wantErr:  "username must be at most 64 characters",
		},
		{
			name:     "weak password",
			username: "testuser",
			password: "weak",
			wantErr:  "password must be at least 8 characters long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Register(nil, tt.username, tt.password)
			if err == nil {
				t.Errorf("Register(%q, %q) expected error, got nil", tt.username, tt.password)
				return
			}
			if err.Error() != tt.wantErr {
				t.Errorf("Register(%q, %q) error = %q, want %q", tt.username, tt.password, err.Error(), tt.wantErr)
			}
		})
	}
}
