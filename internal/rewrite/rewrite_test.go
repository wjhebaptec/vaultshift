package rewrite_test

import (
	"testing"

	"github.com/vaultshift/internal/rewrite"
)

func TestApply_ReplaceKeyPrefix(t *testing.T) {
	r := rewrite.New(rewrite.ReplaceKeyPrefix("prod/", "staging/"))
	k, v, err := r.Apply("prod/db_pass", "secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if k != "staging/db_pass" {
		t.Errorf("expected staging/db_pass, got %q", k)
	}
	if v != "secret" {
		t.Errorf("expected value unchanged, got %q", v)
	}
}

func TestApply_NoMatchingPrefix(t *testing.T) {
	r := rewrite.New(rewrite.ReplaceKeyPrefix("prod/", "staging/"))
	k, _, err := r.Apply("dev/key", "val")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if k != "dev/key" {
		t.Errorf("expected key unchanged, got %q", k)
	}
}

func TestApply_UpperCaseKey(t *testing.T) {
	r := rewrite.New(rewrite.UpperCaseKey())
	k, _, err := r.Apply("my_secret", "val")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if k != "MY_SECRET" {
		t.Errorf("expected MY_SECRET, got %q", k)
	}
}

func TestApply_LowerCaseKey(t *testing.T) {
	r := rewrite.New(rewrite.LowerCaseKey())
	k, _, err := r.Apply("DB_HOST", "localhost")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if k != "db_host" {
		t.Errorf("expected db_host, got %q", k)
	}
}

func TestApply_AppendKeySuffix(t *testing.T) {
	r := rewrite.New(rewrite.AppendKeySuffix("_v2"))
	k, _, err := r.Apply("api_key", "abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if k != "api_key_v2" {
		t.Errorf("expected api_key_v2, got %q", k)
	}
}

func TestApply_TrimValueSpace(t *testing.T) {
	r := rewrite.New(rewrite.TrimValueSpace())
	_, v, err := r.Apply("key", "  hello  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "hello" {
		t.Errorf("expected 'hello', got %q", v)
	}
}

func TestApply_ChainedRules(t *testing.T) {
	r := rewrite.New(
		rewrite.ReplaceKeyPrefix("prod/", ""),
		rewrite.UpperCaseKey(),
		rewrite.TrimValueSpace(),
	)
	k, v, err := r.Apply("prod/db_pass", "  topsecret  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if k != "DB_PASS" {
		t.Errorf("expected DB_PASS, got %q", k)
	}
	if v != "topsecret" {
		t.Errorf("expected topsecret, got %q", v)
	}
}

func TestApplyAll_RewritesAllEntries(t *testing.T) {
	r := rewrite.New(rewrite.UpperCaseKey())
	input := map[string]string{"alpha": "1", "beta": "2"}
	out, err := r.ApplyAll(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, k := range []string{"ALPHA", "BETA"} {
		if _, ok := out[k]; !ok {
			t.Errorf("expected key %q in output", k)
		}
	}
}

func TestApplyAll_PropagatesError(t *testing.T) {
	failRule := func(key, value string) (string, string, error) {
		return "", "", fmt.Errorf("intentional failure")
	}
	r := rewrite.New(failRule)
	_, err := r.ApplyAll(map[string]string{"k": "v"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
