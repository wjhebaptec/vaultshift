// Package label provides key-value labelling for secrets across providers.
// Labels are arbitrary metadata attached to secret keys, enabling grouping,
// filtering, and classification without modifying secret values.
package label

import (
	"errors"
	"fmt"
	"sync"
)

// ErrEmptyKey is returned when a secret key or label key is empty.
var ErrEmptyKey = errors.New("label: key must not be empty")

// ErrEmptyLabelKey is returned when a label key is empty.
var ErrEmptyLabelKey = errors.New("label: label key must not be empty")

// Labels is a map of label key-value pairs.
type Labels map[string]string

// Manager stores and retrieves labels for secret keys scoped by provider.
type Manager struct {
	mu   sync.RWMutex
	data map[string]Labels // store key: "provider:secretKey"
}

// New returns a new Manager.
func New() *Manager {
	return &Manager{
		data: make(map[string]Labels),
	}
}

func storeKey(provider, secretKey string) string {
	return fmt.Sprintf("%s:%s", provider, secretKey)
}

// Set attaches labels to a secret key under the given provider.
// Existing labels are merged; new keys overwrite old ones.
func (m *Manager) Set(provider, secretKey string, labels Labels) error {
	if provider == "" || secretKey == "" {
		return ErrEmptyKey
	}
	for k := range labels {
		if k == "" {
			return ErrEmptyLabelKey
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	sk := storeKey(provider, secretKey)
	if m.data[sk] == nil {
		m.data[sk] = make(Labels)
	}
	for k, v := range labels {
		m.data[sk][k] = v
	}
	return nil
}

// Get returns all labels for a secret key under the given provider.
// Returns nil if no labels are set.
func (m *Manager) Get(provider, secretKey string) (Labels, error) {
	if provider == "" || secretKey == "" {
		return nil, ErrEmptyKey
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	sk := storeKey(provider, secretKey)
	l, ok := m.data[sk]
	if !ok {
		return nil, nil
	}
	copy := make(Labels, len(l))
	for k, v := range l {
		copy[k] = v
	}
	return copy, nil
}

// Delete removes all labels for a secret key under the given provider.
func (m *Manager) Delete(provider, secretKey string) error {
	if provider == "" || secretKey == "" {
		return ErrEmptyKey
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, storeKey(provider, secretKey))
	return nil
}

// Match returns true if the secret key under the given provider has all
// of the supplied label key-value pairs.
func (m *Manager) Match(provider, secretKey string, selector Labels) bool {
	l, err := m.Get(provider, secretKey)
	if err != nil || l == nil {
		return false
	}
	for k, v := range selector {
		if l[k] != v {
			return false
		}
	}
	return true
}
