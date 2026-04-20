package mirror_test

import (
	"context"
	"testing"

	"github.com/vaultshift/internal/mirror"
	"github.com/vaultshift/internal/provider/mock"
)

func TestSync_CopiesAllKeys(t *testing.T) {
	ctx := context.Background()
	primary := mock.New()
	secondary := mock.New()
	m, _ := mirror.New(primary, secondary)

	_ = primary.Put(ctx, "k1", "v1")
	_ = primary.Put(ctx, "k2", "v2")

	res, err := m.Sync(ctx)
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if len(res.Copied) != 2 {
		t.Fatalf("want 2 copied, got %d", len(res.Copied))
	}
	if res.HasFailures() {
		t.Fatalf("unexpected failures: %v", res.Errors)
	}

	for _, k := range []string{"k1", "k2"} {
		if _, err := secondary.Get(ctx, k); err != nil {
			t.Fatalf("secondary missing key %q after sync", k)
		}
	}
}

func TestSync_EmptyPrimary_NoCopied(t *testing.T) {
	ctx := context.Background()
	m, _ := mirror.New(mock.New(), mock.New())

	res, err := m.Sync(ctx)
	if err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if len(res.Copied) != 0 {
		t.Fatalf("want 0 copied, got %d", len(res.Copied))
	}
}

func TestSync_HasFailures_OnSecondaryError(t *testing.T) {
	ctx := context.Background()
	primary := mock.New()
	secondary := mock.New()
	m, _ := mirror.New(primary, secondary)

	_ = primary.Put(ctx, "good", "val")
	_ = primary.Put(ctx, "bad", "val")

	// Close secondary to force write errors.
	secondary.Close()

	res, err := m.Sync(ctx)
	if err != nil {
		t.Fatalf("Sync returned unexpected top-level error: %v", err)
	}
	if !res.HasFailures() {
		t.Fatal("expected failures when secondary is closed")
	}
}

func TestSyncResult_HasFailures_False(t *testing.T) {
	r := &mirror.SyncResult{Copied: []string{"a"}}
	if r.HasFailures() {
		t.Fatal("expected no failures")
	}
}
