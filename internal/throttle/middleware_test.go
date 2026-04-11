package throttle_test

import (
	"context"
	"errors"
	"testing"

	"github.com/vaultshift/internal/throttle"
)

func TestWrapGet_AllowsUnderLimit(t *testing.T) {
	th := throttle.New(throttle.WithRate(5))
	called := false

	fn := throttle.WrapGet(th, "aws", func(ctx context.Context, key string) (string, error) {
		called = true
		return "secret-value", nil
	})

	val, err := fn(context.Background(), "my/secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "secret-value" {
		t.Fatalf("expected 'secret-value', got %q", val)
	}
	if !called {
		t.Fatal("inner function was not called")
	}
}

func TestWrapGet_BlocksOverLimit(t *testing.T) {
	th := throttle.New(throttle.WithRate(1))
	ctx := context.Background()

	fn := throttle.WrapGet(th, "gcp", func(ctx context.Context, key string) (string, error) {
		return "v", nil
	})

	_, _ = fn(ctx, "k")
	_, err := fn(ctx, "k")
	if err == nil {
		t.Fatal("expected throttle error on second call")
	}
}

func TestWrapPut_AllowsUnderLimit(t *testing.T) {
	th := throttle.New(throttle.WithRate(5))
	called := false

	fn := throttle.WrapPut(th, "vault", func(ctx context.Context, key, value string) error {
		called = true
		return nil
	})

	if err := fn(context.Background(), "my/secret", "newval"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("inner function was not called")
	}
}

func TestWrapPut_PropagatesInnerError(t *testing.T) {
	th := throttle.New(throttle.WithRate(5))
	expected := errors.New("write failed")

	fn := throttle.WrapPut(th, "aws", func(ctx context.Context, key, value string) error {
		return expected
	})

	err := fn(context.Background(), "k", "v")
	if !errors.Is(err, expected) {
		t.Fatalf("expected inner error, got: %v", err)
	}
}

func TestWrapGet_EmptyScope_UsesKeyDirectly(t *testing.T) {
	th := throttle.New(throttle.WithRate(2))

	fn := throttle.WrapGet(th, "", func(ctx context.Context, key string) (string, error) {
		return "ok", nil
	})

	_, err := fn(context.Background(), "bare-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := th.Usage("bare-key"); got != 1 {
		t.Fatalf("expected usage 1, got %d", got)
	}
}
