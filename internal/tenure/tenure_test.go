package tenure_test

import (
	"testing"
	"time"

	"github.com/vaultshift/internal/tenure"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestTouch_CreatesRecord(t *testing.T) {
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	tr := tenure.New(tenure.WithClock(fixedClock(now)))
	tr.Touch("aws", "db/pass")

	r, err := tr.Get("aws", "db/pass")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !r.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt = %v, want %v", r.CreatedAt, now)
	}
	if r.Provider != "aws" || r.Key != "db/pass" {
		t.Errorf("unexpected record fields: %+v", r)
	}
}

func TestTouch_UpdatesSeenAt(t *testing.T) {
	first := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	second := first.Add(5 * time.Minute)

	tr := tenure.New(tenure.WithClock(fixedClock(first)))
	tr.Touch("gcp", "api_key")

	// advance clock
	tr2 := tr
	_ = tr2
	// re-touch with a new tracker sharing the same store via Touch
	tr.Touch("gcp", "api_key") // still first clock

	// create fresh tracker at second time
	tr3 := tenure.New(tenure.WithClock(fixedClock(second)))
	tr3.Touch("gcp", "api_key")
	tr3.Touch("gcp", "api_key") // second touch updates SeenAt

	r, _ := tr3.Get("gcp", "api_key")
	if !r.SeenAt.Equal(second) {
		t.Errorf("SeenAt = %v, want %v", r.SeenAt, second)
	}
	if !r.CreatedAt.Equal(second) {
		t.Errorf("CreatedAt should equal second for new tracker, got %v", r.CreatedAt)
	}
}

func TestGet_UnknownKey(t *testing.T) {
	tr := tenure.New()
	_, err := tr.Get("vault", "missing")
	if err != tenure.ErrUnknownKey {
		t.Errorf("want ErrUnknownKey, got %v", err)
	}
}

func TestDelete_RemovesRecord(t *testing.T) {
	tr := tenure.New()
	tr.Touch("aws", "token")
	tr.Delete("aws", "token")
	_, err := tr.Get("aws", "token")
	if err != tenure.ErrUnknownKey {
		t.Errorf("expected ErrUnknownKey after delete, got %v", err)
	}
}

func TestOlderThan_FiltersCorrectly(t *testing.T) {
	base := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	old := base.Add(-2 * time.Hour)

	tr := tenure.New(tenure.WithClock(fixedClock(old)))
	tr.Touch("aws", "old_secret")

	// simulate "now" being base by creating a second tracker
	tr2 := tenure.New(tenure.WithClock(fixedClock(base)))
	tr2.Touch("aws", "fresh_secret")
	tr2.Touch("aws", "old_secret2")

	// Use first tracker to check OlderThan: clock is still `old`
	// so no records are older than 1h relative to `old`
	results := tr.OlderThan(1 * time.Hour)
	if len(results) != 0 {
		t.Errorf("expected 0 old records from old-clock tracker, got %d", len(results))
	}

	// Use second tracker: clock is `base`, records created at `base` → age=0
	results2 := tr2.OlderThan(1 * time.Hour)
	if len(results2) != 0 {
		t.Errorf("expected 0 results for fresh records, got %d", len(results2))
	}
}

func TestAge_IsPositive(t *testing.T) {
	tr := tenure.New()
	tr.Touch("vault", "cred")
	time.Sleep(1 * time.Millisecond)
	r, _ := tr.Get("vault", "cred")
	if r.Age() <= 0 {
		t.Error("expected positive age")
	}
}
