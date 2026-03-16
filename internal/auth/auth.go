package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cli-auth/internal/lockout"
	"github.com/cli-auth/internal/session"
	"github.com/cli-auth/internal/user"
)

// Service orchestrates authentication operations.
type Service struct {
	users    *user.Store
	sessions *session.Store
	lockout  *lockout.Service
}

// NewService creates a new auth service.
func NewService(users *user.Store, sessions *session.Store, lockout *lockout.Service) *Service {
	return &Service{
		users:    users,
		sessions: sessions,
		lockout:  lockout,
	}
}

// LoginResult contains the result of a successful login.
type LoginResult struct {
	User      *user.User
	Token     string
	ExpiresAt time.Time
}

// Register creates a new user account.
func (s *Service) Register(ctx context.Context, username, password string) (*user.User, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}
	if len(username) < 3 {
		return nil, fmt.Errorf("username must be at least 3 characters")
	}
	if len(username) > 64 {
		return nil, fmt.Errorf("username must be at most 64 characters")
	}
	if len(password) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters")
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	u, err := s.users.Create(ctx, username, hash)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return nil, fmt.Errorf("username '%s' is already taken", username)
		}
		return nil, err
	}
	return u, nil
}

// Login authenticates a user with username and password, optionally validates TOTP.
// If TOTP is enabled and totpCode is empty, returns ErrTOTPRequired.
func (s *Service) Login(ctx context.Context, username, password, totpCode string) (*LoginResult, error) {
	u, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	// Check lockout
	locked, remaining, err := s.lockout.Check(ctx, u)
	if err != nil {
		return nil, err
	}
	if locked {
		mins := int(remaining.Minutes()) + 1
		return nil, fmt.Errorf("account is locked. Try again in %d minute(s)", mins)
	}

	// Verify password
	if !CheckPassword(u.PasswordHash, password) {
		lockErr := s.lockout.RecordFailure(ctx, u.ID)
		if lockErr != nil {
			return nil, lockErr
		}
		return nil, fmt.Errorf("invalid username or password")
	}

	// Check TOTP if enabled
	if u.TOTPEnabled {
		if totpCode == "" {
			return nil, &ErrTOTPRequired{UserID: u.ID, Username: u.Username}
		}
		if u.TOTPSecret == nil || !ValidateCode(*u.TOTPSecret, totpCode) {
			lockErr := s.lockout.RecordFailure(ctx, u.ID)
			if lockErr != nil {
				return nil, lockErr
			}
			return nil, fmt.Errorf("invalid TOTP code")
		}
	}

	// Success — reset lockout, record login, create session
	if err := s.lockout.Reset(ctx, u.ID); err != nil {
		return nil, err
	}
	if err := s.users.RecordLogin(ctx, u.ID); err != nil {
		return nil, err
	}

	token, err := s.sessions.Create(ctx, u.ID)
	if err != nil {
		return nil, err
	}

	// Refresh user to get updated last_login_at
	u, err = s.users.GetByID(ctx, u.ID)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		User:      u,
		Token:     token,
		ExpiresAt: time.Now().UTC().Add(s.sessions.Timeout()),
	}, nil
}

// Logout invalidates a session.
func (s *Service) Logout(ctx context.Context, token string) error {
	return s.sessions.Delete(ctx, token)
}

// ValidateSession checks if the current session is still valid.
func (s *Service) ValidateSession(ctx context.Context, token string) (*user.User, *session.Session, error) {
	sess, err := s.sessions.Validate(ctx, token)
	if err != nil {
		return nil, nil, err
	}
	u, err := s.users.GetByID(ctx, sess.UserID)
	if err != nil {
		return nil, nil, err
	}
	return u, sess, nil
}

// EnableTOTP generates a new TOTP secret for a user.
// Returns the OTP key for QR code generation. The caller must then verify the code
// and call ConfirmTOTP to actually enable it.
func (s *Service) EnableTOTP(username string) (*TOTPEnrollment, error) {
	key, err := GenerateSecret(username)
	if err != nil {
		return nil, err
	}
	return &TOTPEnrollment{
		Key:    key,
		Secret: key.Secret(),
		URL:    key.URL(),
	}, nil
}

// ConfirmTOTP verifies a TOTP code and enables 2FA for the user.
func (s *Service) ConfirmTOTP(ctx context.Context, userID int, secret, code string) error {
	if !ValidateCode(secret, code) {
		return fmt.Errorf("invalid TOTP code — 2FA not enabled")
	}
	return s.users.UpdateTOTP(ctx, userID, secret, true)
}

// DisableTOTP disables 2FA after verifying the current TOTP code.
func (s *Service) DisableTOTP(ctx context.Context, userID int, secret, code string) error {
	if !ValidateCode(secret, code) {
		return fmt.Errorf("invalid TOTP code — 2FA not disabled")
	}
	return s.users.UpdateTOTP(ctx, userID, "", false)
}

// TOTPEnrollment holds the data needed for TOTP enrollment.
type TOTPEnrollment struct {
	Key    interface{ URL() string }
	Secret string
	URL    string
}

// ErrTOTPRequired is returned when login requires a TOTP code.
type ErrTOTPRequired struct {
	UserID   int
	Username string
}

func (e *ErrTOTPRequired) Error() string {
	return "TOTP code required"
}
