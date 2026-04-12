package expire

import (
	"testing"
	"time"
)

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestRegister_AndGet(t *testing.T) {
	tracker := New()
	base := time.Now()
	tracker.now = fixedNow(base)

	tracker.Register("aws", "db/password", 24*time.Hour)
	e, ok := tracker.Get("aws", "db/password")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Key != "db/password" || e.Provider != "aws" {
		t.Errorf("unexpected entry: %+v", e)
	}
	if !e.ExpiresAt.Equal(base.Add(24 * time.Hour)) {
		t.Errorf("unexpected ExpiresAt: %v", e.ExpiresAt)
	}
}

func TestGet_UnknownKey(t *testing.T) {
	tracker := New()
	_, ok := tracker.Get("aws", "missing")
	if ok {
		t.Fatal("expected no entry for unknown key")
	}
}

func TestIsExpired_True(t *testing.T) {
	base := time.Now()
	tracker := New()
	tracker.now = fixedNow(base)
	tracker.Register("gcp", "api/key", time.Second)

	tracker.now = fixedNow(base.Add(2 * time.Second))
	expired := tracker.Expired()
	if len(expired) != 1 || expired[0].Key != "api/key" {
		t.Errorf("expected 1 expired entry, got %v", expired)
	}
}

func TestIsExpired_False(t *testing.T) {
	base := time.Now()
	tracker := New()
	tracker.now = fixedNow(base)
	tracker.Register("gcp", "api/key", time.Hour)

	expired := tracker.Expired()
	if len(expired) != 0 {
		t.Errorf("expected no expired entries, got %v", expired)
	}
}

func TestExpiringSoon_WithinWindow(t *testing.T) {
	base := time.Now()
	tracker := New()
	tracker.now = fixedNow(base)
	tracker.Register("vault", "secret/token", 5*time.Minute)
	tracker.Register("vault", "secret/cert", 2*time.Hour)

	soon := tracker.ExpiringSoon(10 * time.Minute)
	if len(soon) != 1 || soon[0].Key != "secret/token" {
		t.Errorf("expected 1 expiring-soon entry, got %v", soon)
	}
}

func TestRemove_DeletesEntry(t *testing.T) {
	tracker := New()
	tracker.Register("aws", "key1", time.Hour)
	tracker.Remove("aws", "key1")
	_, ok := tracker.Get("aws", "key1")
	if ok {
		t.Fatal("expected entry to be removed")
	}
}

func TestExpiresIn_ReturnsZeroWhenExpired(t *testing.T) {
	base := time.Now()
	e := Entry{ExpiresAt: base.Add(-time.Second)}
	if d := e.ExpiresIn(base); d != 0 {
		t.Errorf("expected 0, got %v", d)
	}
}

func TestExpired_MultipleProviders_Independent(t *testing.T) {
	base := time.Now()
	tracker := New()
	tracker.now = fixedNow(base)
	tracker.Register("aws", "k", time.Millisecond)
	tracker.Register("gcp", "k", time.Hour)

	tracker.now = fixedNow(base.Add(10 * time.Millisecond))
	expired := tracker.Expired()
	if len(expired) != 1 || expired[0].Provider != "aws" {
		t.Errorf("expected only aws/k to be expired, got %v", expired)
	}
}
