package tag_test

import (
	"errors"
	"sort"
	"testing"

	"github.com/vaultshift/internal/tag"
)

func TestSet_AndGet(t *testing.T) {
	s := tag.New()
	if err := s.Set("mykey", "env", "staging"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, err := s.Get("mykey", "env")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "staging" {
		t.Errorf("expected staging, got %q", v)
	}
}

func TestGet_UnknownSecret(t *testing.T) {
	s := tag.New()
	_, err := s.Get("nope", "env")
	if !errors.Is(err, tag.ErrTagNotFound) {
		t.Errorf("expected ErrTagNotFound, got %v", err)
	}
}

func TestGet_UnknownTagKey(t *testing.T) {
	s := tag.New()
	_ = s.Set("mykey", "env", "prod")
	_, err := s.Get("mykey", "missing")
	if !errors.Is(err, tag.ErrTagNotFound) {
		t.Errorf("expected ErrTagNotFound, got %v", err)
	}
}

func TestSet_EmptySecretKey(t *testing.T) {
	s := tag.New()
	if err := s.Set("", "env", "prod"); err == nil {
		t.Error("expected error for empty secret key")
	}
}

func TestSet_EmptyTagKey(t *testing.T) {
	s := tag.New()
	if err := s.Set("mykey", "", "prod"); err == nil {
		t.Error("expected error for empty tag key")
	}
}

func TestTags_ReturnsCopy(t *testing.T) {
	s := tag.New()
	_ = s.Set("mykey", "env", "prod")
	_ = s.Set("mykey", "team", "platform")
	tags := s.Tags("mykey")
	if len(tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(tags))
	}
	// mutating the copy must not affect the store
	tags["env"] = "mutated"
	v, _ := s.Get("mykey", "env")
	if v != "prod" {
		t.Errorf("store was mutated via returned map")
	}
}

func TestTags_UnknownSecret(t *testing.T) {
	s := tag.New()
	if tags := s.Tags("ghost"); tags != nil {
		t.Errorf("expected nil for unknown secret, got %v", tags)
	}
}

func TestDelete_RemovesTag(t *testing.T) {
	s := tag.New()
	_ = s.Set("mykey", "env", "prod")
	s.Delete("mykey", "env")
	_, err := s.Get("mykey", "env")
	if !errors.Is(err, tag.ErrTagNotFound) {
		t.Errorf("expected ErrTagNotFound after delete, got %v", err)
	}
}

func TestMatchAll_ReturnsMatchingKeys(t *testing.T) {
	s := tag.New()
	_ = s.Set("db/pass", "env", "prod")
	_ = s.Set("db/pass", "team", "platform")
	_ = s.Set("api/key", "env", "prod")
	_ = s.Set("api/key", "team", "backend")
	_ = s.Set("dev/secret", "env", "dev")

	matches := s.MatchAll(map[string]string{"env": "prod", "team": "platform"})
	if len(matches) != 1 || matches[0] != "db/pass" {
		t.Errorf("unexpected matches: %v", matches)
	}
}

func TestMatchAll_MultipleResults(t *testing.T) {
	s := tag.New()
	_ = s.Set("a", "env", "prod")
	_ = s.Set("b", "env", "prod")
	_ = s.Set("c", "env", "dev")

	matches := s.MatchAll(map[string]string{"env": "prod"})
	sort.Strings(matches)
	if len(matches) != 2 || matches[0] != "a" || matches[1] != "b" {
		t.Errorf("unexpected matches: %v", matches)
	}
}
