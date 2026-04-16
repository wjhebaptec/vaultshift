package stamp_test

import (
	"testing"
	"time"

	"github.com/vaultshift/internal/stamp"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestAttach_AndExtract_RoundTrip(t *testing.T) {
	now := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	s, _ := stamp.New(stamp.WithClock(fixedClock(now)))
	stamped := s.Attach("mysecret")
	value, ts, err := s.Extract(stamped)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != "mysecret" {
		t.Errorf("got value %q, want %q", value, "mysecret")
	}
	if !ts.Equal(now) {
		t.Errorf("got ts %v, want %v", ts, now)
	}
}

func TestExtract_NoTimestamp_ReturnsError(t *testing.T) {
	s, _ := stamp.New()
	_, _, err := s.Extract("plainvalue")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestExtract_InvalidTimestamp_ReturnsError(t *testing.T) {
	s, _ := stamp.New()
	_, _, err := s.Extract("value|notadate")
	if err == nil {
		t.Fatal("expected error for invalid timestamp")
	}
}

func TestAge_ReturnsCorrectDuration(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	s, _ := stamp.New(stamp.WithClock(fixedClock(base)))
	stamped := s.Attach("v")

	later := base.Add(5 * time.Minute)
	s2, _ := stamp.New(stamp.WithClock(fixedClock(later)))
	age, err := s2.Age(stamped)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if age != 5*time.Minute {
		t.Errorf("got age %v, want 5m", age)
	}
}

func TestNew_EmptySeparator_ReturnsError(t *testing.T) {
	_, err := stamp.New(stamp.WithSeparator(""))
	if err == nil {
		t.Fatal("expected error for empty separator")
	}
}

func TestAttach_CustomSeparator(t *testing.T) {
	now := time.Date(2024, 3, 15, 8, 0, 0, 0, time.UTC)
	s, _ := stamp.New(stamp.WithSeparator("::"), stamp.WithClock(fixedClock(now)))
	stamped := s.Attach("data")
	value, _, err := s.Extract(stamped)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != "data" {
		t.Errorf("got %q, want %q", value, "data")
	}
}
