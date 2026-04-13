package overwrite_test

import (
	"context"
	"errors"
	"testing"

	"github.com/vaultshift/internal/overwrite"
	"github.com/vaultshift/internal/provider/mock"
)

func setup(t *testing.T) (*overwrite.Guard, *mock.Provider) {
	t.Helper()
	mp := mock.New()
	g, err := overwrite.New(mp)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return g, mp
}

func TestNew_NilProvider_ReturnsError(t *testing.T) {
	_, err := overwrite.New(nil)
	if !errors.Is(err, overwrite.ErrNilProvider) {
		t.Fatalf("expected ErrNilProvider, got %v", err)
	}
}

func TestPut_WritesWhenAbsent(t *testing.T) {
	g, _ := setup(t)
	ctx := context.Background()
	r, err := g.Put(ctx, "db/pass", "secret")
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	if !r.Written {
		t.Error("expected Written=true for new key")
	}
}

func TestPut_SkipsWhenValueUnchanged(t *testing.T) {
	g, _ := setup(t)
	ctx := context.Background()
	if _, err := g.Put(ctx, "db/pass", "secret"); err != nil {
		t.Fatalf("first Put: %v", err)
	}
	r, err := g.Put(ctx, "db/pass", "secret")
	if err != nil {
		t.Fatalf("second Put: %v", err)
	}
	if r.Written {
		t.Error("expected Written=false when value is unchanged")
	}
}

func TestPut_WritesWhenValueChanged(t *testing.T) {
	g, _ := setup(t)
	ctx := context.Background()
	if _, err := g.Put(ctx, "db/pass", "old"); err != nil {
		t.Fatalf("first Put: %v", err)
	}
	r, err := g.Put(ctx, "db/pass", "new")
	if err != nil {
		t.Fatalf("second Put: %v", err)
	}
	if !r.Written {
		t.Error("expected Written=true when value changed")
	}
}

func TestGet_DelegatesToInner(t *testing.T) {
	g, mp := setup(t)
	ctx := context.Background()
	_ = mp.Put(ctx, "k", "v")
	val, err := g.Get(ctx, "k")
	if err != nil || val != "v" {
		t.Fatalf("Get: got %q, %v", val, err)
	}
}

func TestPutAll_WritesOnlyChanged(t *testing.T) {
	g, _ := setup(t)
	ctx := context.Background()
	// seed one key
	if _, err := g.Put(ctx, "a", "1"); err != nil {
		t.Fatalf("seed: %v", err)
	}
	results, err := g.PutAll(ctx, map[string]string{"a": "1", "b": "2"})
	if err != nil {
		t.Fatalf("PutAll: %v", err)
	}
	written := 0
	for _, r := range results {
		if r.Written {
			written++
		}
	}
	if written != 1 {
		t.Errorf("expected 1 write, got %d", written)
	}
}
