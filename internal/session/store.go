package session

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"
)

// Store handles database operations for sessions.
type Store struct {
	db              *sql.DB
	timeout         time.Duration
	sessionFilePath string
}

// NewStore creates a new session store.
func NewStore(db *sql.DB, timeout time.Duration, sessionFilePath string) *Store {
	return &Store{
		db:              db,
		timeout:         timeout,
		sessionFilePath: sessionFilePath,
	}
}

// Create generates a new session token, stores it in the DB, and writes it to the session file.
func (s *Store) Create(ctx context.Context, userID int) (string, error) {
	token, err := generateToken()
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().UTC().Add(s.timeout)

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO sessions (user_id, token, expires_at) VALUES ($1, $2, $3)`,
		userID, token, expiresAt,
	)
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}

	// Clean up any expired sessions for this user
	_, _ = s.db.ExecContext(ctx,
		`DELETE FROM sessions WHERE user_id = $1 AND expires_at < NOW()`, userID)

	// Write token to session file for persistence
	if err := s.writeSessionFile(token); err != nil {
		return "", fmt.Errorf("write session file: %w", err)
	}

	return token, nil
}

// Validate checks if a token is valid and not expired.
func (s *Store) Validate(ctx context.Context, token string) (*Session, error) {
	sess := &Session{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, user_id, token, created_at, expires_at 
		 FROM sessions 
		 WHERE token = $1 AND expires_at > NOW()`,
		token,
	).Scan(&sess.ID, &sess.UserID, &sess.Token, &sess.CreatedAt, &sess.ExpiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session expired or not found")
		}
		return nil, fmt.Errorf("validate session: %w", err)
	}
	return sess, nil
}

// Delete removes a session from the DB and deletes the session file.
func (s *Store) Delete(ctx context.Context, token string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE token = $1`, token)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	// Remove session file
	_ = os.Remove(s.sessionFilePath)

	return nil
}

// LoadToken reads the session token from the session file.
func (s *Store) LoadToken() (string, error) {
	data, err := os.ReadFile(s.sessionFilePath)
	if err != nil {
		return "", fmt.Errorf("read session file: %w", err)
	}
	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", fmt.Errorf("empty session file")
	}
	return token, nil
}

// Timeout returns the configured session timeout duration.
func (s *Store) Timeout() time.Duration {
	return s.timeout
}

// writeSessionFile writes the token to the session file.
func (s *Store) writeSessionFile(token string) error {
	return os.WriteFile(s.sessionFilePath, []byte(token), 0600)
}

// generateToken creates a cryptographically secure random token.
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return hex.EncodeToString(b), nil
}
