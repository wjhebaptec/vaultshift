// Package cursor provides pagination cursor support for listing secrets
// across providers, enabling resumable and bounded list operations.
package cursor

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
)

// ErrInvalidCursor is returned when a cursor token cannot be decoded.
var ErrInvalidCursor = errors.New("cursor: invalid token")

// Cursor holds the pagination state for a list operation.
type Cursor struct {
	Provider string `json:"provider"`
	Offset   int    `json:"offset"`
	Prefix   string `json:"prefix,omitempty"`
	PageSize int    `json:"page_size"`
}

// Manager encodes and decodes pagination cursors.
type Manager struct {
	defaultPageSize int
}

// Option configures a Manager.
type Option func(*Manager)

// WithPageSize sets the default page size.
func WithPageSize(n int) Option {
	return func(m *Manager) {
		if n > 0 {
			m.defaultPageSize = n
		}
	}
}

// New creates a Manager with the given options.
func New(opts ...Option) *Manager {
	m := &Manager{defaultPageSize: 50}
	for _, o := range opts {
		o(m)
	}
	return m
}

// Encode serialises a Cursor into an opaque base64 token.
func (m *Manager) Encode(c Cursor) (string, error) {
	if c.Provider == "" {
		return "", fmt.Errorf("cursor: provider must not be empty")
	}
	if c.PageSize <= 0 {
		c.PageSize = m.defaultPageSize
	}
	b, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("cursor: encode: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Decode parses an opaque token back into a Cursor.
func (m *Manager) Decode(token string) (Cursor, error) {
	if token == "" {
		return Cursor{}, ErrInvalidCursor
	}
	b, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return Cursor{}, ErrInvalidCursor
	}
	var c Cursor
	if err := json.Unmarshal(b, &c); err != nil {
		return Cursor{}, ErrInvalidCursor
	}
	if c.Provider == "" {
		return Cursor{}, ErrInvalidCursor
	}
	return c, nil
}

// Next returns a cursor advanced by one page.
func (m *Manager) Next(c Cursor) Cursor {
	n := c
	if n.PageSize <= 0 {
		n.PageSize = m.defaultPageSize
	}
	n.Offset += n.PageSize
	return n
}
