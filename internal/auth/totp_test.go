package auth

import (
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
)

func TestGenerateSecret(t *testing.T) {
	key, err := GenerateSecret("testuser")
	if err != nil {
		t.Fatalf("GenerateSecret() unexpected error: %v", err)
	}

	if key == nil {
		t.Fatal("GenerateSecret() returned nil key")
	}

	if key.Secret() == "" {
		t.Error("GenerateSecret() key has empty secret")
	}

	url := key.URL()
	if url == "" {
		t.Error("GenerateSecret() key has empty URL")
	}

	// URL should contain the issuer and account name
	if key.Issuer() != totpIssuer {
		t.Errorf("GenerateSecret() issuer = %q, want %q", key.Issuer(), totpIssuer)
	}
	if key.AccountName() != "testuser" {
		t.Errorf("GenerateSecret() account = %q, want %q", key.AccountName(), "testuser")
	}
}

func TestGenerateSecret_UniqueSecrets(t *testing.T) {
	key1, err := GenerateSecret("user1")
	if err != nil {
		t.Fatalf("GenerateSecret(user1) error: %v", err)
	}

	key2, err := GenerateSecret("user2")
	if err != nil {
		t.Fatalf("GenerateSecret(user2) error: %v", err)
	}

	if key1.Secret() == key2.Secret() {
		t.Error("GenerateSecret() produced identical secrets for different users")
	}
}

func TestValidateCode(t *testing.T) {
	// Generate a real secret, then generate a valid code for it
	key, err := GenerateSecret("testuser")
	if err != nil {
		t.Fatalf("GenerateSecret() error: %v", err)
	}
	secret := key.Secret()

	// Generate a valid code using the library
	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		t.Fatalf("totp.GenerateCode() error: %v", err)
	}

	if !ValidateCode(secret, code) {
		t.Error("ValidateCode() returned false for valid code")
	}

	// Invalid code
	if ValidateCode(secret, "000000") && ValidateCode(secret, "999999") {
		// Extremely unlikely both would be valid at the same time
		t.Error("ValidateCode() returned true for likely-invalid codes")
	}

	// Wrong secret
	key2, _ := GenerateSecret("other")
	if ValidateCode(key2.Secret(), code) {
		t.Error("ValidateCode() returned true for wrong secret")
	}
}

func TestValidateCode_InvalidInputs(t *testing.T) {
	if ValidateCode("", "123456") {
		t.Error("ValidateCode() returned true for empty secret")
	}

	key, _ := GenerateSecret("testuser")
	if ValidateCode(key.Secret(), "") {
		t.Error("ValidateCode() returned true for empty code")
	}
}
