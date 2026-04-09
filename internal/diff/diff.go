// Package diff provides utilities for comparing secrets between providers.
package diff

import "fmt"

// ChangeType represents the kind of change detected between two secret stores.
type ChangeType string

const (
	ChangeAdded   ChangeType = "added"
	ChangeRemoved ChangeType = "removed"
	ChangeUpdated ChangeType = "updated"
	ChangeUnchanged ChangeType = "unchanged"
)

// Change describes a single secret difference.
type Change struct {
	Key        string
	ChangeType ChangeType
	OldValue   string
	NewValue   string
}

// String returns a human-readable description of the change.
func (c Change) String() string {
	switch c.ChangeType {
	case ChangeAdded:
		return fmt.Sprintf("[+] %s", c.Key)
	case ChangeRemoved:
		return fmt.Sprintf("[-] %s", c.Key)
	case ChangeUpdated:
		return fmt.Sprintf("[~] %s", c.Key)
	default:
		return fmt.Sprintf("[=] %s", c.Key)
	}
}

// Compare computes the diff between a source and destination map of secrets.
// It returns a slice of Change entries describing additions, removals, and updates.
func Compare(src, dst map[string]string) []Change {
	var changes []Change

	for key, srcVal := range src {
		dstVal, exists := dst[key]
		if !exists {
			changes = append(changes, Change{
				Key:        key,
				ChangeType: ChangeAdded,
				NewValue:   srcVal,
			})
			continue
		}
		if srcVal != dstVal {
			changes = append(changes, Change{
				Key:        key,
				ChangeType: ChangeUpdated,
				OldValue:   dstVal,
				NewValue:   srcVal,
			})
		} else {
			changes = append(changes, Change{
				Key:        key,
				ChangeType: ChangeUnchanged,
				OldValue:   dstVal,
				NewValue:   srcVal,
			})
		}
	}

	for key, dstVal := range dst {
		if _, exists := src[key]; !exists {
			changes = append(changes, Change{
				Key:        key,
				ChangeType: ChangeRemoved,
				OldValue:   dstVal,
			})
		}
	}

	return changes
}

// HasDrift returns true if any added, removed, or updated changes exist.
func HasDrift(changes []Change) bool {
	for _, c := range changes {
		if c.ChangeType != ChangeUnchanged {
			return true
		}
	}
	return false
}
