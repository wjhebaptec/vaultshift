package redact_test

import (
	"testing"

	"github.com/vaultshift/internal/redact"
)

func TestRedact_RegisteredValueIsReplaced(t *testing.T) {
	r := redact.New()
	r.Register("supersecret")

	got := r.Redact("the password is supersecret, keep it safe")
	want := "the password is [REDACTED], keep it safe"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRedact_UnregisteredValueIsPreserved(t *testing.T) {
	r := redact.New()
	input := "nothing sensitive here"
	if got := r.Redact(input); got != input {
		t.Errorf("expected input unchanged, got %q", got)
	}
}

func TestRedact_EmptyValueIgnored(t *testing.T) {
	r := redact.New()
	r.Register("") // should be a no-op
	got := r.Redact("hello")
	if got != "hello" {
		t.Errorf("unexpected redaction: %q", got)
	}
}

func TestRedact_CustomPlaceholder(t *testing.T) {
	r := redact.New(redact.WithPlaceholder("***"))
	r.Register("tok123")
	got := r.Redact("token: tok123")
	if got != "token: ***" {
		t.Errorf("got %q", got)
	}
}

func TestRedact_MultipleOccurrences(t *testing.T) {
	r := redact.New()
	r.Register("abc")
	got := r.Redact("abc and abc again")
	if got != "[REDACTED] and [REDACTED] again" {
		t.Errorf("got %q", got)
	}
}

func TestForget_StopsRedacting(t *testing.T) {
	r := redact.New()
	r.Register("secret")
	r.Forget("secret")
	got := r.Redact("my secret value")
	if got != "my secret value" {
		t.Errorf("expected no redaction after Forget, got %q", got)
	}
}

func TestRedactMap_RedactsValues(t *testing.T) {
	r := redact.New()
	r.Register("p@ssw0rd")

	input := map[string]string{
		"db_pass": "p@ssw0rd",
		"db_user": "admin",
	}
	out := r.RedactMap(input)

	if out["db_pass"] != "[REDACTED]" {
		t.Errorf("db_pass not redacted: %q", out["db_pass"])
	}
	if out["db_user"] != "admin" {
		t.Errorf("db_user should be unchanged, got %q", out["db_user"])
	}
}

func TestRedactMap_DoesNotMutateOriginal(t *testing.T) {
	r := redact.New()
	r.Register("secret")

	orig := map[string]string{"key": "secret"}
	_ = r.RedactMap(orig)

	if orig["key"] != "secret" {
		t.Error("original map was mutated")
	}
}
