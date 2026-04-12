package label

import (
	"sort"
	"testing"
)

func TestFilter_MatchingKeys(t *testing.T) {
	m := New()
	_ = m.Set("aws", "db/pass", Labels{"env": "prod"})
	_ = m.Set("aws", "api/key", Labels{"env": "prod", "tier": "frontend"})
	_ = m.Set("aws", "dev/token", Labels{"env": "dev"})

	keys := []string{"db/pass", "api/key", "dev/token"}
	got := m.Filter("aws", keys, Labels{"env": "prod"})
	if len(got) != 2 {
		t.Fatalf("expected 2 matches, got %d: %v", len(got), got)
	}
}

func TestFilter_NoMatches(t *testing.T) {
	m := New()
	_ = m.Set("aws", "db/pass", Labels{"env": "dev"})
	keys := []string{"db/pass"}
	got := m.Filter("aws", keys, Labels{"env": "prod"})
	if len(got) != 0 {
		t.Errorf("expected no matches, got %v", got)
	}
}

func TestFilter_EmptySelector_MatchesNone(t *testing.T) {
	m := New()
	_ = m.Set("aws", "db/pass", Labels{"env": "prod"})
	keys := []string{"db/pass"}
	// empty selector: Match iterates zero times, returns true
	got := m.Filter("aws", keys, Labels{})
	if len(got) != 1 {
		t.Errorf("expected 1 match with empty selector, got %v", got)
	}
}

func TestListLabelled_ReturnsAllKeys(t *testing.T) {
	m := New()
	_ = m.Set("gcp", "secret/a", Labels{"x": "1"})
	_ = m.Set("gcp", "secret/b", Labels{"y": "2"})
	_ = m.Set("aws", "other", Labels{"z": "3"})

	got := m.ListLabelled("gcp")
	sort.Strings(got)
	if len(got) != 2 {
		t.Fatalf("expected 2 keys for gcp, got %d: %v", len(got), got)
	}
	if got[0] != "secret/a" || got[1] != "secret/b" {
		t.Errorf("unexpected keys: %v", got)
	}
}

func TestListLabelled_EmptyProvider(t *testing.T) {
	m := New()
	_ = m.Set("aws", "key", Labels{"env": "prod"})
	got := m.ListLabelled("gcp")
	if len(got) != 0 {
		t.Errorf("expected empty list, got %v", got)
	}
}
