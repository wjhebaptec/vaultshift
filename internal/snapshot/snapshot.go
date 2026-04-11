// Package snapshot provides point-in-time captures of secret state
// across one or more providers, enabling comparison and restoration.
package snapshot

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Provider is the minimal interface required to capture a snapshot.
type Provider interface {
	List(ctx context.Context) ([]string, error)
	Get(ctx context.Context, key string) (string, error)
}

// Snapshot holds a named, timestamped copy of secrets from a provider.
type Snapshot struct {
	Name      string
	Provider  string
	CapturedAt time.Time
	Secrets   map[string]string
}

// Manager captures and stores snapshots in memory.
type Manager struct {
	mu        sync.RWMutex
	snapshots map[string]*Snapshot
}

// New returns a new snapshot Manager.
func New() *Manager {
	return &Manager{
		snapshots: make(map[string]*Snapshot),
	}
}

// Capture reads all secrets from the provider and stores them under name.
func (m *Manager) Capture(ctx context.Context, name, providerName string, p Provider) (*Snapshot, error) {
	keys, err := p.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("snapshot: list keys from %q: %w", providerName, err)
	}

	secrets := make(map[string]string, len(keys))
	for _, k := range keys {
		v, err := p.Get(ctx, k)
		if err != nil {
			return nil, fmt.Errorf("snapshot: get key %q from %q: %w", k, providerName, err)
		}
		secrets[k] = v
	}

	snap := &Snapshot{
		Name:       name,
		Provider:   providerName,
		CapturedAt: time.Now().UTC(),
		Secrets:    secrets,
	}

	m.mu.Lock()
	m.snapshots[name] = snap
	m.mu.Unlock()

	return snap, nil
}

// Get retrieves a previously captured snapshot by name.
func (m *Manager) Get(name string) (*Snapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snap, ok := m.snapshots[name]
	if !ok {
		return nil, fmt.Errorf("snapshot: %q not found", name)
	}
	return snap, nil
}

// Delete removes a snapshot by name.
func (m *Manager) Delete(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.snapshots, name)
}

// List returns the names of all stored snapshots.
func (m *Manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.snapshots))
	for n := range m.snapshots {
		names = append(names, n)
	}
	return names
}
