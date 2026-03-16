package session

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestGenerateToken(t *testing.T) {
	token, err := generateToken()
	if err != nil {
		t.Fatalf("generateToken() unexpected error: %v", err)
	}

	// Token should be a 64-character hex string (32 bytes)
	if len(token) != 64 {
		t.Errorf("generateToken() length = %d, want 64", len(token))
	}

	// Should only contain hex characters
	for _, ch := range token {
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f')) {
			t.Errorf("generateToken() contains non-hex character: %c", ch)
			break
		}
	}
}

func TestGenerateToken_Unique(t *testing.T) {
	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token, err := generateToken()
		if err != nil {
			t.Fatalf("generateToken() error on iteration %d: %v", i, err)
		}
		if tokens[token] {
			t.Fatalf("generateToken() produced duplicate token on iteration %d", i)
		}
		tokens[token] = true
	}
}

func TestWriteAndLoadSessionFile(t *testing.T) {
	tmpDir := t.TempDir()
	sessionFile := filepath.Join(tmpDir, ".session")

	store := NewStore(nil, 30*time.Minute, sessionFile)

	// Write a token
	testToken := "abc123def456"
	err := store.writeSessionFile(testToken)
	if err != nil {
		t.Fatalf("writeSessionFile() error: %v", err)
	}

	// Verify file permissions (0600) - skip on Windows as it doesn't enforce Unix-style permissions
	if runtime.GOOS != "windows" {
		info, err := os.Stat(sessionFile)
		if err != nil {
			t.Fatalf("os.Stat() error: %v", err)
		}
		perm := info.Mode().Perm()
		if perm != 0600 {
			t.Errorf("session file permissions = %o, want 0600", perm)
		}
	}

	// Load the token back
	loaded, err := store.LoadToken()
	if err != nil {
		t.Fatalf("LoadToken() error: %v", err)
	}
	if loaded != testToken {
		t.Errorf("LoadToken() = %q, want %q", loaded, testToken)
	}
}

func TestLoadToken_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	sessionFile := filepath.Join(tmpDir, "nonexistent")

	store := NewStore(nil, 30*time.Minute, sessionFile)

	_, err := store.LoadToken()
	if err == nil {
		t.Fatal("LoadToken() expected error for missing file, got nil")
	}
}

func TestLoadToken_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	sessionFile := filepath.Join(tmpDir, ".session")

	// Create an empty file
	if err := os.WriteFile(sessionFile, []byte(""), 0600); err != nil {
		t.Fatalf("setup: write empty file error: %v", err)
	}

	store := NewStore(nil, 30*time.Minute, sessionFile)

	_, err := store.LoadToken()
	if err == nil {
		t.Fatal("LoadToken() expected error for empty file, got nil")
	}
	if !strings.Contains(err.Error(), "empty session file") {
		t.Errorf("LoadToken() error = %q, want to contain 'empty session file'", err.Error())
	}
}

func TestLoadToken_WhitespaceHandling(t *testing.T) {
	tmpDir := t.TempDir()
	sessionFile := filepath.Join(tmpDir, ".session")

	// Token with trailing newline (common from file writes)
	if err := os.WriteFile(sessionFile, []byte("  mytoken123  \n"), 0600); err != nil {
		t.Fatalf("setup: write file error: %v", err)
	}

	store := NewStore(nil, 30*time.Minute, sessionFile)

	token, err := store.LoadToken()
	if err != nil {
		t.Fatalf("LoadToken() error: %v", err)
	}
	if token != "mytoken123" {
		t.Errorf("LoadToken() = %q, want %q", token, "mytoken123")
	}
}

func TestTimeout(t *testing.T) {
	timeout := 45 * time.Minute
	store := NewStore(nil, timeout, "/tmp/.session")

	if store.Timeout() != timeout {
		t.Errorf("Timeout() = %v, want %v", store.Timeout(), timeout)
	}
}

func TestSession_Struct(t *testing.T) {
	now := time.Now()
	expires := now.Add(30 * time.Minute)

	sess := &Session{
		ID:        "test-uuid",
		UserID:    42,
		Token:     "test-token",
		CreatedAt: now,
		ExpiresAt: expires,
	}

	if sess.ID != "test-uuid" {
		t.Errorf("Session.ID = %q, want %q", sess.ID, "test-uuid")
	}
	if sess.UserID != 42 {
		t.Errorf("Session.UserID = %d, want 42", sess.UserID)
	}
	if sess.Token != "test-token" {
		t.Errorf("Session.Token = %q, want %q", sess.Token, "test-token")
	}
	if !sess.ExpiresAt.After(sess.CreatedAt) {
		t.Error("Session.ExpiresAt should be after CreatedAt")
	}
}
