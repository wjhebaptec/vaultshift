// Package tag provides utilities for attaching and querying metadata tags
// on secrets, enabling filtering, grouping, and policy enforcement by label.
package tag

import (
	"errors"
	"fmt"
	"sync"
)

// ErrTagNotFound is returned when a requested tag key does not exist.
var ErrTagNotFound = errors.New("tag not found")

// Store holds tags associated with named secrets.
type Store struct {
	mu   sync.RWMutex
	data map[string]map[string]string // secret key -> tag key -> tag value
}

// New creates a new empty tag Store.
func New() *Store {
	return &Store{
		data: make(map[string]map[string]string),
	}
}

// Set attaches a tag key/value pair to the given secret key.
func (s *Store) Set(secretKey, tagKey, tagValue string) error {
	if secretKey == "" {
		return errors.New("secret key must not be empty")
	}
	if tagKey == "" {
		return errors.New("tag key must not be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[secretKey]; !ok {
		s.data[secretKey] = make(map[string]string)
	}
	s.data[secretKey][tagKey] = tagValue
	return nil
}

// Get retrieves the value of a tag key for the given secret key.
func (s *Store) Get(secretKey, tagKey string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tags, ok := s.data[secretKey]
	if !ok {
		return "", fmt.Errorf("%w: secret %q has no tags", ErrTagNotFound, secretKey)
	}
	v, ok := tags[tagKey]
	if !ok {
		return "", fmt.Errorf("%w: key %q on secret %q", ErrTagNotFound, tagKey, secretKey)
	}
	return v, nil
}

// Tags returns a copy of all tags for the given secret key.
func (s *Store) Tags(secretKey string) map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src, ok := s.data[secretKey]
	if !ok {
		return nil
	}
	copy := make(map[string]string, len(src))
	for k, v := range src {
		copy[k] = v
	}
	return copy
}

// Delete removes a single tag key from the given secret key.
func (s *Store) Delete(secretKey, tagKey string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if tags, ok := s.data[secretKey]; ok {
		delete(tags, tagKey)
	}
}

// MatchAll returns all secret keys whose tags contain all of the provided
// key/value pairs.
func (s *Store) MatchAll(filter map[string]string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var results []string
outer:
	for secretKey, tags := range s.data {
		for fk, fv := range filter {
			if tags[fk] != fv {
				continue outer
			}
		}
		results = append(results, secretKey)
	}
	return results
}
