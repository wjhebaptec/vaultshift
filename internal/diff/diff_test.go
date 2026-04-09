package diff_test

import (
	"testing"

	"github.com/vaultshift/internal/diff"
)

func TestCompare_Added(t *testing.T) {
	src := map[string]string{"KEY_A": "val1", "KEY_B": "val2"}
	dst := map[string]string{"KEY_A": "val1"}

	changes := diff.Compare(src, dst)
	found := findChange(changes, "KEY_B")
	if found == nil {
		t.Fatal("expected KEY_B to appear as added")
	}
	if found.ChangeType != diff.ChangeAdded {
		t.Errorf("expected added, got %s", found.ChangeType)
	}
}

func TestCompare_Removed(t *testing.T) {
	src := map[string]string{"KEY_A": "val1"}
	dst := map[string]string{"KEY_A": "val1", "KEY_B": "old"}

	changes := diff.Compare(src, dst)
	found := findChange(changes, "KEY_B")
	if found == nil {
		t.Fatal("expected KEY_B to appear as removed")
	}
	if found.ChangeType != diff.ChangeRemoved {
		t.Errorf("expected removed, got %s", found.ChangeType)
	}
}

func TestCompare_Updated(t *testing.T) {
	src := map[string]string{"KEY_A": "new_val"}
	dst := map[string]string{"KEY_A": "old_val"}

	changes := diff.Compare(src, dst)
	found := findChange(changes, "KEY_A")
	if found == nil {
		t.Fatal("expected KEY_A to appear as updated")
	}
	if found.ChangeType != diff.ChangeUpdated {
		t.Errorf("expected updated, got %s", found.ChangeType)
	}
	if found.OldValue != "old_val" || found.NewValue != "new_val" {
		t.Errorf("unexpected values: old=%s new=%s", found.OldValue, found.NewValue)
	}
}

func TestCompare_Unchanged(t *testing.T) {
	src := map[string]string{"KEY_A": "same"}
	dst := map[string]string{"KEY_A": "same"}

	changes := diff.Compare(src, dst)
	if diff.HasDrift(changes) {
		t.Error("expected no drift for identical maps")
	}
}

func TestHasDrift_True(t *testing.T) {
	src := map[string]string{"NEW": "val"}
	dst := map[string]string{}
	if !diff.HasDrift(diff.Compare(src, dst)) {
		t.Error("expected drift to be detected")
	}
}

func TestChange_String(t *testing.T) {
	cases := []struct {
		ct   diff.ChangeType
		want string
	}{
		{diff.ChangeAdded, "[+] mykey"},
		{diff.ChangeRemoved, "[-] mykey"},
		{diff.ChangeUpdated, "[~] mykey"},
		{diff.ChangeUnchanged, "[=] mykey"},
	}
	for _, tc := range cases {
		c := diff.Change{Key: "mykey", ChangeType: tc.ct}
		if c.String() != tc.want {
			t.Errorf("got %q, want %q", c.String(), tc.want)
		}
	}
}

func findChange(changes []diff.Change, key string) *diff.Change {
	for i := range changes {
		if changes[i].Key == key {
			return &changes[i]
		}
	}
	return nil
}
