package auth

import (
	"fmt"
	"io"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/mdp/qrterminal/v3"
)

const totpIssuer = "cli-auth"

// GenerateSecret creates a new TOTP secret key for a user.
func GenerateSecret(username string) (*otp.Key, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      totpIssuer,
		AccountName: username,
	})
	if err != nil {
		return nil, fmt.Errorf("generate totp secret: %w", err)
	}
	return key, nil
}

// RenderQR prints a QR code for the TOTP key to the given writer.
func RenderQR(w io.Writer, key *otp.Key) {
	config := qrterminal.Config{
		Level:     qrterminal.L,
		Writer:    w,
		BlackChar: qrterminal.BLACK,
		WhiteChar: qrterminal.WHITE,
		QuietZone: 1,
	}
	qrterminal.GenerateWithConfig(key.URL(), config)
}

// ValidateCode validates a 6-digit TOTP code against a secret.
func ValidateCode(secret, code string) bool {
	return totp.Validate(code, secret)
}
