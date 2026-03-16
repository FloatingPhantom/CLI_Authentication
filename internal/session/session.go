package session

import "time"

// Session represents an active user session.
type Session struct {
	ID        string
	UserID    int
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
}
