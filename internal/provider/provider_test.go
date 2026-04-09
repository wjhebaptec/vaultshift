package provider

import (
	"context"
	"testing"
)

// mockProvider is a test implementation of Provider
type mockProvider struct {
	providerType Type
	secrets      map[string]*Secret
	closed       bool
}

func newMockProvider(t Type) *mockProvider {
	return &mockProvider{
		providerType: t,
		secrets:      make(map[string]*Secret),
	}
}

func (m *mockProvider) GetSecret(ctx context.Context, key string) (*Secret, error) {
	s, ok := m.secrets[key]
	if !ok {
		return nil, nil
	}
	return s, nil
}

func (m *mockProvider) PutSecret(ctx context.Context, secret *Secret) error {
	m.secrets[secret.Key] = secret
	return nil
}

func (m *mockProvider) DeleteSecret(ctx context.Context, key string) error {
	delete(m.secrets, key)
	return nil
}

func (m *mockProvider) ListSecrets(ctx context.Context) ([]string, error) {
	keys := make([]string, 0, len(m.secrets))
	for k := range m.secrets {
		keys = append(keys, k)
	}
	return keys, nil
}

func (m *mockProvider) Type() Type {
	return m.providerType
}

func (m *mockProvider) Close() error {
	m.closed = true
	return nil
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	registry := NewRegistry()
	awsProvider := newMockProvider(TypeAWS)
	
	registry.Register(awsProvider)
	
	p, err := registry.Get(TypeAWS)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	
	if p.Type() != TypeAWS {
		t.Errorf("expected type %s, got %s", TypeAWS, p.Type())
	}
}

func TestRegistry_GetNonExistent(t *testing.T) {
	registry := NewRegistry()
	
	_, err := registry.Get(TypeGCP)
	if err == nil {
		t.Error("expected error for non-existent provider")
	}
}

func TestRegistry_Close(t *testing.T) {
	registry := NewRegistry()
	awsProvider := newMockProvider(TypeAWS)
	gcpProvider := newMockProvider(TypeGCP)
	
	registry.Register(awsProvider)
	registry.Register(gcpProvider)
	
	if err := registry.Close(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	
	if !awsProvider.closed || !gcpProvider.closed {
		t.Error("expected all providers to be closed")
	}
}
