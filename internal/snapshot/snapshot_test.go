package snapshot_test

import (
	"context"
	"errors"
	"testing"

	"github.com/vaultshift/internal/snapshot"
)

// stubProvider is a simple in-memory provider for testing.
type stubProvider struct {
	secrets map[string]string
	listErr error
	getErr  error
}

func (s *stubProvider) List(_ context.Context) ([]string, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	keys := make([]string, 0, len(s.secrets))
	for k := range s.secrets {
		keys = append(keys, k)
	}
	return keys, nil
}

func (s *stubProvider) Get(_ context.Context, key string) (string, error) {
	if s.getErr != nil {
		return "", s.getErr
	}
	v, ok := s.secrets[key]
	if !ok {
		return "", errors.New("not found")
	}
	return v, nil
}

func TestCapture_StoresSnapshot(t *testing.T) {
	m := snapshot.New()
	p := &stubProvider{secrets: map[string]string{"db/pass": "s3cr3t", "api/key": "abc123"}}

	snap, err := m.Capture(context.Background(), "snap1", "aws", p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap.Name != "snap1" || snap.Provider != "aws" {
		t.Fatalf("unexpected snapshot metadata: %+v", snap)
	}
	if snap.Secrets["db/pass"] != "s3cr3t" {
		t.Errorf("expected secret value, got %q", snap.Secrets["db/pass"])
	}
	if snap.CapturedAt.IsZero() {
		t.Error("expected CapturedAt to be set")
	}
}

func TestGet_ReturnsStoredSnapshot(t *testing.T) {
	m := snapshot.New()
	p := &stubProvider{secrets: map[string]string{"k": "v"}}
	m.Capture(context.Background(), "mysnap", "gcp", p) //nolint

	snap, err := m.Get("mysnap")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap.Secrets["k"] != "v" {
		t.Errorf("expected v, got %q", snap.Secrets["k"])
	}
}

func TestGet_NotFound(t *testing.T) {
	m := snapshot.New()
	_, err := m.Get("ghost")
	if err == nil {
		t.Fatal("expected error for missing snapshot")
	}
}

func TestDelete_RemovesSnapshot(t *testing.T) {
	m := snapshot.New()
	p := &stubProvider{secrets: map[string]string{"x": "y"}}
	m.Capture(context.Background(), "tmp", "vault", p) //nolint
	m.Delete("tmp")

	if _, err := m.Get("tmp"); err == nil {
		t.Fatal("expected snapshot to be deleted")
	}
}

func TestCapture_ListError(t *testing.T) {
	m := snapshot.New()
	p := &stubProvider{listErr: errors.New("list failed")}
	_, err := m.Capture(context.Background(), "bad", "aws", p)
	if err == nil {
		t.Fatal("expected error from list failure")
	}
}

func TestList_ReturnsAllNames(t *testing.T) {
	m := snapshot.New()
	p := &stubProvider{secrets: map[string]string{"a": "1"}}
	m.Capture(context.Background(), "s1", "aws", p) //nolint
	m.Capture(context.Background(), "s2", "gcp", p) //nolint

	names := m.List()
	if len(names) != 2 {
		t.Errorf("expected 2 snapshots, got %d", len(names))
	}
}
