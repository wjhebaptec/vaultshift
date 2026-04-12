package admit_test

import (
	"context"
	"errors"
	"testing"

	"github.com/vaultshift/internal/admit"
)

func TestRegister_AddsPolicy(t *testing.T) {
	a := admit.New()
	err := a.Register("allow-all", func(_ context.Context, _ admit.Request) error { return nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Len() != 1 {
		t.Fatalf("expected 1 policy, got %d", a.Len())
	}
}

func TestRegister_DuplicateName_ReturnsError(t *testing.T) {
	a := admit.New()
	fn := func(_ context.Context, _ admit.Request) error { return nil }
	_ = a.Register("p", fn)
	err := a.Register("p", fn)
	if err == nil {
		t.Fatal("expected error for duplicate policy name")
	}
}

func TestRegister_EmptyName_ReturnsError(t *testing.T) {
	a := admit.New()
	err := a.Register("", func(_ context.Context, _ admit.Request) error { return nil })
	if err == nil {
		t.Fatal("expected error for empty policy name")
	}
}

func TestRegister_NilFunc_ReturnsError(t *testing.T) {
	a := admit.New()
	err := a.Register("nil-fn", nil)
	if err == nil {
		t.Fatal("expected error for nil policy function")
	}
}

func TestAdmit_AllowsWhenNoPolicies(t *testing.T) {
	a := admit.New()
	req := admit.Request{Provider: "aws", Key: "db/pass", Op: admit.OpGet}
	if err := a.Admit(context.Background(), req); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestAdmit_DeniesOnFirstFailingPolicy(t *testing.T) {
	a := admit.New()
	_ = a.Register("allow", func(_ context.Context, _ admit.Request) error { return nil })
	_ = a.Register("deny", admit.DenyAll)

	err := a.Admit(context.Background(), admit.Request{Provider: "aws", Key: "x", Op: admit.OpGet})
	if err == nil {
		t.Fatal("expected denial, got nil")
	}
	if !errors.Is(err, admit.ErrDenied) {
		t.Fatalf("expected ErrDenied, got %v", err)
	}
}

func TestAllowReadOnly_PermitsReads(t *testing.T) {
	a := admit.New()
	_ = a.Register("ro", admit.AllowReadOnly)

	for _, op := range []admit.Op{admit.OpGet, admit.OpList} {
		req := admit.Request{Provider: "gcp", Key: "k", Op: op}
		if err := a.Admit(context.Background(), req); err != nil {
			t.Errorf("op %q should be allowed, got %v", op, err)
		}
	}
}

func TestAllowReadOnly_DeniesWrites(t *testing.T) {
	a := admit.New()
	_ = a.Register("ro", admit.AllowReadOnly)

	for _, op := range []admit.Op{admit.OpPut, admit.OpDelete} {
		req := admit.Request{Provider: "gcp", Key: "k", Op: op}
		err := a.Admit(context.Background(), req)
		if !errors.Is(err, admit.ErrDenied) {
			t.Errorf("op %q should be denied, got %v", op, err)
		}
	}
}
