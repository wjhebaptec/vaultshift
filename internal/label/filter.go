package label

// Filter returns the subset of secret keys (from the provided slice) whose
// labels match all entries in selector, scoped to the given provider.
func (m *Manager) Filter(provider string, keys []string, selector Labels) []string {
	var matched []string
	for _, k := range keys {
		if m.Match(provider, k, selector) {
			matched = append(matched, k)
		}
	}
	return matched
}

// ListLabelled returns all secret keys under the given provider that have at
// least one label set.
func (m *Manager) ListLabelled(provider string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	prefix := provider + ":"
	var keys []string
	for k := range m.data {
		if len(k) > len(prefix) && k[:len(prefix)] == prefix {
			keys = append(keys, k[len(prefix):])
		}
	}
	return keys
}
