package lockout

import (
	"context"
	"fmt"
	"time"

	"github.com/cli-auth/internal/user"
)

const (
	// MaxAttempts before account lockout.
	MaxAttempts = 5
	// LockoutDuration is how long the account stays locked.
	LockoutDuration = 15 * time.Minute
)

// Service handles account lockout logic.
type Service struct {
	users *user.Store
}

// NewService creates a new lockout service.
func NewService(users *user.Store) *Service {
	return &Service{users: users}
}

// Check verifies if the user account is currently locked.
// Returns (isLocked, remainingDuration, error).
func (s *Service) Check(ctx context.Context, u *user.User) (bool, time.Duration, error) {
	if u.LockedUntil == nil {
		return false, 0, nil
	}
	remaining := time.Until(*u.LockedUntil)
	if remaining <= 0 {
		// Lock expired — reset
		if err := s.users.ResetFailedAttempts(ctx, u.ID); err != nil {
			return false, 0, err
		}
		return false, 0, nil
	}
	return true, remaining, nil
}

// RecordFailure increments failed attempts and locks the account if threshold is reached.
// Returns an error describing the lockout state if the account gets locked.
func (s *Service) RecordFailure(ctx context.Context, userID int) error {
	count, err := s.users.IncrementFailedAttempts(ctx, userID)
	if err != nil {
		return err
	}

	if count >= MaxAttempts {
		until := time.Now().UTC().Add(LockoutDuration)
		if err := s.users.LockAccount(ctx, userID, until); err != nil {
			return err
		}
		return fmt.Errorf("account locked for %v due to %d failed attempts", LockoutDuration, count)
	}

	remaining := MaxAttempts - count
	return fmt.Errorf("invalid credentials (%d attempt(s) remaining before lockout)", remaining)
}

// Reset clears failed attempts and lock status after a successful login.
func (s *Service) Reset(ctx context.Context, userID int) error {
	return s.users.ResetFailedAttempts(ctx, userID)
}
