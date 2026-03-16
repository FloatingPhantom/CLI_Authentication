package lockout

import (
	"testing"
	"time"

	"github.com/cli-auth/internal/user"
)

func TestCheck_NotLocked(t *testing.T) {
	// User with no lock — should not be locked
	u := &user.User{
		ID:          1,
		Username:    "testuser",
		LockedUntil: nil,
	}

	svc := &Service{users: nil} // users won't be called for this path
	locked, remaining, err := svc.Check(nil, u)
	if err != nil {
		t.Fatalf("Check() unexpected error: %v", err)
	}
	if locked {
		t.Error("Check() returned locked=true for user with no lock")
	}
	if remaining != 0 {
		t.Errorf("Check() remaining = %v, want 0", remaining)
	}
}

func TestCheck_StillLocked(t *testing.T) {
	// User locked until 1 hour from now
	future := time.Now().Add(1 * time.Hour)
	u := &user.User{
		ID:          1,
		Username:    "testuser",
		LockedUntil: &future,
	}

	svc := &Service{users: nil} // users won't be called for this path
	locked, remaining, err := svc.Check(nil, u)
	if err != nil {
		t.Fatalf("Check() unexpected error: %v", err)
	}
	if !locked {
		t.Error("Check() returned locked=false for user with future lock")
	}
	if remaining <= 0 {
		t.Errorf("Check() remaining = %v, want > 0", remaining)
	}
}

func TestConstants(t *testing.T) {
	if MaxAttempts != 5 {
		t.Errorf("MaxAttempts = %d, want 5", MaxAttempts)
	}
	if LockoutDuration != 15*time.Minute {
		t.Errorf("LockoutDuration = %v, want 15m", LockoutDuration)
	}
}
