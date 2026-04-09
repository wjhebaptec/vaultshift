package lock_test

import (
	"testing"
	"time"

	"github.com/vaultshift/internal/lock"
)

func TestAcquire_Success(t *testing.T) {
	m := lock.New()
	if err := m.Acquire("db/password", "worker-1", 0); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestAcquire_AlreadyLocked(t *testing.T) {
	m := lock.New()
	_ = m.Acquire("db/password", "worker-1", 0)

	err := m.Acquire("db/password", "worker-2", 0)
	if err != lock.ErrAlreadyLocked {
		t.Fatalf("expected ErrAlreadyLocked, got %v", err)
	}
}

func TestRelease_Success(t *testing.T) {
	m := lock.New()
	_ = m.Acquire("db/password", "worker-1", 0)

	if err := m.Release("db/password"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if m.IsLocked("db/password") {
		t.Fatal("expected key to be unlocked after release")
	}
}

func TestRelease_NotLocked(t *testing.T) {
	m := lock.New()
	err := m.Release("db/password")
	if err != lock.ErrNotLocked {
		t.Fatalf("expected ErrNotLocked, got %v", err)
	}
}

func TestIsLocked_False(t *testing.T) {
	m := lock.New()
	if m.IsLocked("nonexistent") {
		t.Fatal("expected unlocked for unknown key")
	}
}

func TestAcquire_AfterExpiry(t *testing.T) {
	m := lock.New()
	_ = m.Acquire("api/key", "worker-1", 10*time.Millisecond)

	time.Sleep(20 * time.Millisecond)

	if err := m.Acquire("api/key", "worker-2", 0); err != nil {
		t.Fatalf("expected lock to be re-acquirable after expiry, got %v", err)
	}
}

func TestGet_ReturnsEntry(t *testing.T) {
	m := lock.New()
	_ = m.Acquire("vault/token", "worker-3", time.Minute)

	e, ok := m.Get("vault/token")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Owner != "worker-3" {
		t.Errorf("expected owner worker-3, got %s", e.Owner)
	}
}

func TestGet_ExpiredReturnsNotFound(t *testing.T) {
	m := lock.New()
	_ = m.Acquire("vault/token", "worker-3", 5*time.Millisecond)

	time.Sleep(15 * time.Millisecond)

	_, ok := m.Get("vault/token")
	if ok {
		t.Fatal("expected expired entry to return not found")
	}
}

func TestIsExpired_NoTTL(t *testing.T) {
	e := lock.Entry{AcquiredAt: time.Now().Add(-24 * time.Hour), TTL: 0}
	if e.IsExpired() {
		t.Fatal("expected entry with zero TTL to never expire")
	}
}
