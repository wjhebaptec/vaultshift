package checkpoint

import (
	"testing"
)

func TestMark_AndGet(t *testing.T) {
	cp := New()
	cp.Mark("db/pass", StatusCompleted, "")

	e, err := cp.Get("db/pass")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Status != StatusCompleted {
		t.Errorf("expected Completed, got %s", e.Status)
	}
	if e.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestGet_UnknownKey(t *testing.T) {
	cp := New()
	_, err := cp.Get("missing")
	if err == nil {
		t.Fatal("expected error for unknown key")
	}
}

func TestMark_RecordsError(t *testing.T) {
	cp := New()
	cp.Mark("api/key", StatusFailed, "provider unavailable")

	e, err := cp.Get("api/key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Error != "provider unavailable" {
		t.Errorf("expected error message, got %q", e.Error)
	}
}

func TestPending_ReturnsPendingAndRunning(t *testing.T) {
	cp := New()
	cp.Mark("k1", StatusPending, "")
	cp.Mark("k2", StatusRunning, "")
	cp.Mark("k3", StatusCompleted, "")
	cp.Mark("k4", StatusFailed, "")

	pending := cp.Pending()
	if len(pending) != 2 {
		t.Errorf("expected 2 pending, got %d", len(pending))
	}
}

func TestReset_ClearsAllEntries(t *testing.T) {
	cp := New()
	cp.Mark("k1", StatusCompleted, "")
	cp.Mark("k2", StatusFailed, "")
	cp.Reset()

	if _, err := cp.Get("k1"); err == nil {
		t.Error("expected error after reset")
	}
}

func TestSummary_CountsByStatus(t *testing.T) {
	cp := New()
	cp.Mark("k1", StatusCompleted, "")
	cp.Mark("k2", StatusCompleted, "")
	cp.Mark("k3", StatusFailed, "")
	cp.Mark("k4", StatusPending, "")

	summary := cp.Summary()
	if summary[StatusCompleted] != 2 {
		t.Errorf("expected 2 completed, got %d", summary[StatusCompleted])
	}
	if summary[StatusFailed] != 1 {
		t.Errorf("expected 1 failed, got %d", summary[StatusFailed])
	}
	if summary[StatusPending] != 1 {
		t.Errorf("expected 1 pending, got %d", summary[StatusPending])
	}
}

func TestMark_OverwritesPreviousEntry(t *testing.T) {
	cp := New()/x", StatusRunning, "")
	cp.Mark("secret/x", StatusCompleted, "")

	e, _ := cp.Get("secret/x")
	if e.Status != StatusCompleted {
		t.Errorf("expected Completed after overwrite, got %s", e.Status)
	}
}
