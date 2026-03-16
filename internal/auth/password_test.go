package auth

import (
	"testing"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  string
	}{
		{
			name:     "valid password",
			password: "Str0ng!Pass",
			wantErr:  "",
		},
		{
			name:     "too short",
			password: "Ab1!",
			wantErr:  "password must be at least 8 characters long",
		},
		{
			name:     "missing uppercase",
			password: "lowercase1!",
			wantErr:  "password must contain at least one uppercase letter",
		},
		{
			name:     "missing lowercase",
			password: "UPPERCASE1!",
			wantErr:  "password must contain at least one lowercase letter",
		},
		{
			name:     "missing digit",
			password: "NoDigits!!",
			wantErr:  "password must contain at least one digit",
		},
		{
			name:     "missing special char",
			password: "NoSpecial1a",
			wantErr:  "password must contain at least one special character",
		},
		{
			name:     "exactly 8 characters valid",
			password: "Abcde1!x",
			wantErr:  "",
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  "password must be at least 8 characters long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePassword(tt.password)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("validatePassword(%q) unexpected error: %v", tt.password, err)
				}
			} else {
				if err == nil {
					t.Errorf("validatePassword(%q) expected error %q, got nil", tt.password, tt.wantErr)
				} else if err.Error() != tt.wantErr {
					t.Errorf("validatePassword(%q) error = %q, want %q", tt.password, err.Error(), tt.wantErr)
				}
			}
		})
	}
}

func TestHashPassword(t *testing.T) {
	password := "Str0ng!Pass"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() unexpected error: %v", err)
	}

	if hash == "" {
		t.Fatal("HashPassword() returned empty hash")
	}

	if hash == password {
		t.Fatal("HashPassword() returned plaintext password")
	}

	// Hashing twice should produce different hashes (bcrypt uses random salt)
	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() second call unexpected error: %v", err)
	}
	if hash == hash2 {
		t.Error("HashPassword() produced identical hashes for same input (expected different salts)")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "Str0ng!Pass"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() setup error: %v", err)
	}

	tests := []struct {
		name     string
		hash     string
		password string
		want     bool
	}{
		{
			name:     "correct password",
			hash:     hash,
			password: password,
			want:     true,
		},
		{
			name:     "wrong password",
			hash:     hash,
			password: "WrongPass1!",
			want:     false,
		},
		{
			name:     "empty password",
			hash:     hash,
			password: "",
			want:     false,
		},
		{
			name:     "invalid hash",
			hash:     "not-a-valid-hash",
			password: password,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckPassword(tt.hash, tt.password)
			if got != tt.want {
				t.Errorf("CheckPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}
