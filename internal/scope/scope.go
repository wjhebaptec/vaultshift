// Package scope provides namespace-scoped access control for secrets,
// restricting provider operations to a defined key prefix boundary.
package scope

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// ErrOutOfScope is returned when a key falls outside the allowed scope.
var ErrOutOfScope = errors.New("key is out of scope")

// Provider is the minimal interface required by the scoped wrapper.
type Provider interface {
	Get(ctx context.Context, key string) (string, error)
	Put(ctx context.Context, key, value string) error
	Delete(ctx context.Context, key string) error
	List(ctx context.Context) ([]string, error)
}

// Scoped wraps a Provider and enforces a namespace boundary.
type Scoped struct {
	inner  Provider
	prefix string
	sep    string
}

// New returns a Scoped provider that restricts all operations to keys
// beginning with the given namespace prefix.
func New(inner Provider, namespace, separator string) (*Scoped, error) {
	if inner == nil {
		return nil, errors.New("inner provider must not be nil")
	}
	if namespace == "" {
		return nil, errors.New("namespace must not be empty")
	}
	if separator == "" {
		separator = "/"
	}
	return &Scoped{inner: inner, prefix: namespace, sep: separator}, nil
}

func (s *Scoped) qualify(key string) string {
	return s.prefix + s.sep + key
}

func (s *Scoped) assertInScope(key string) error {
	expected := s.prefix + s.sep
	if !strings.HasPrefix(key, expected) {
		return fmt.Errorf("%w: %q does not start with %q", ErrOutOfScope, key, expected)
	}
	return nil
}

// Get retrieves a secret, qualifying the key with the namespace prefix.
func (s *Scoped) Get(ctx context.Context, key string) (string, error) {
	return s.inner.Get(ctx, s.qualify(key))
}

// Put stores a secret under the namespaced key.
func (s *Scoped) Put(ctx context.Context, key, value string) error {
	return s.inner.Put(ctx, s.qualify(key), value)
}

// Delete removes a secret identified by the namespaced key.
func (s *Scoped) Delete(ctx context.Context, key string) error {
	return s.inner.Delete(ctx, s.qualify(key))
}

// List returns only the keys that belong to this scope, stripping the prefix.
func (s *Scoped) List(ctx context.Context) ([]string, error) {
	all, err := s.inner.List(ctx)
	if err != nil {
		return nil, err
	}
	expected := s.prefix + s.sep
	var result []string
	for _, k := range all {
		if strings.HasPrefix(k, expected) {
			result = append(result, strings.TrimPrefix(k, expected))
		}
	}
	return result, nil
}

// InScope reports whether the given raw (unqualified) key would be within scope
// after qualification. It can also be used to validate pre-qualified keys.
func (s *Scoped) InScope(key string) bool {
	return s.assertInScope(s.qualify(key)) == nil
}
