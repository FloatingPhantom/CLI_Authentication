package user

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Store handles database operations for users.
type Store struct {
	db *sql.DB
}

// NewStore creates a new user store.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// Create inserts a new user and returns it.
func (s *Store) Create(ctx context.Context, username, passwordHash string) (*User, error) {
	u := &User{}
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO users (username, password_hash) 
		 VALUES ($1, $2) 
		 RETURNING id, username, password_hash, totp_secret, totp_enabled,
		           failed_attempts, locked_until, created_at, last_login_at`,
		username, passwordHash,
	).Scan(
		&u.ID, &u.Username, &u.PasswordHash,
		&u.TOTPSecret, &u.TOTPEnabled,
		&u.FailedAttempts, &u.LockedUntil,
		&u.CreatedAt, &u.LastLoginAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

// GetByUsername fetches a user by username.
func (s *Store) GetByUsername(ctx context.Context, username string) (*User, error) {
	u := &User{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, username, password_hash, totp_secret, totp_enabled,
		        failed_attempts, locked_until, created_at, last_login_at
		 FROM users WHERE username = $1`,
		username,
	).Scan(
		&u.ID, &u.Username, &u.PasswordHash,
		&u.TOTPSecret, &u.TOTPEnabled,
		&u.FailedAttempts, &u.LockedUntil,
		&u.CreatedAt, &u.LastLoginAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	return u, nil
}

// GetByID fetches a user by ID.
func (s *Store) GetByID(ctx context.Context, id int) (*User, error) {
	u := &User{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, username, password_hash, totp_secret, totp_enabled,
		        failed_attempts, locked_until, created_at, last_login_at
		 FROM users WHERE id = $1`,
		id,
	).Scan(
		&u.ID, &u.Username, &u.PasswordHash,
		&u.TOTPSecret, &u.TOTPEnabled,
		&u.FailedAttempts, &u.LockedUntil,
		&u.CreatedAt, &u.LastLoginAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	return u, nil
}

// UpdateTOTP sets the TOTP secret and enabled flag.
func (s *Store) UpdateTOTP(ctx context.Context, userID int, secret string, enabled bool) error {
	var secretVal *string
	if enabled {
		secretVal = &secret
	}
	_, err := s.db.ExecContext(ctx,
		`UPDATE users SET totp_secret = $1, totp_enabled = $2 WHERE id = $3`,
		secretVal, enabled, userID,
	)
	if err != nil {
		return fmt.Errorf("update totp: %w", err)
	}
	return nil
}

// RecordLogin sets last_login_at to now.
func (s *Store) RecordLogin(ctx context.Context, userID int) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE users SET last_login_at = $1 WHERE id = $2`,
		time.Now().UTC(), userID,
	)
	if err != nil {
		return fmt.Errorf("record login: %w", err)
	}
	return nil
}

// IncrementFailedAttempts increments failed_attempts and returns the new count.
func (s *Store) IncrementFailedAttempts(ctx context.Context, userID int) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`UPDATE users SET failed_attempts = failed_attempts + 1 
		 WHERE id = $1 
		 RETURNING failed_attempts`,
		userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("increment failed attempts: %w", err)
	}
	return count, nil
}

// LockAccount sets the locked_until timestamp.
func (s *Store) LockAccount(ctx context.Context, userID int, until time.Time) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE users SET locked_until = $1 WHERE id = $2`,
		until, userID,
	)
	if err != nil {
		return fmt.Errorf("lock account: %w", err)
	}
	return nil
}

// ResetFailedAttempts resets the failed_attempts counter and clears lock.
func (s *Store) ResetFailedAttempts(ctx context.Context, userID int) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE users SET failed_attempts = 0, locked_until = NULL WHERE id = $1`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("reset failed attempts: %w", err)
	}
	return nil
}
