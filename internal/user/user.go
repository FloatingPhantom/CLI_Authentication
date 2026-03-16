package user

import "time"

// User represents a registered user in the system.
type User struct {
	ID             int
	Username       string
	PasswordHash   string
	TOTPSecret     *string
	TOTPEnabled    bool
	FailedAttempts int
	LockedUntil    *time.Time
	CreatedAt      time.Time
	LastLoginAt    *time.Time
}
