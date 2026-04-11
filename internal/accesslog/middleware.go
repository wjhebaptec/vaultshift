package accesslog

import (
	"context"
	"time"
)

// Provider is a minimal interface matching the vaultshift provider contract.
type Provider interface {
	Get(ctx context.Context, key string) (string, error)
	Put(ctx context.Context, key, value string) error
	Delete(ctx context.Context, key string) error
	List(ctx context.Context) ([]string, error)
}

// Middleware wraps a Provider and records each access in the given Log.
type Middleware struct {
	provider Provider
	log      *Log
	name     string
}

// Wrap returns a new Middleware that records accesses for the named provider.
func Wrap(name string, p Provider, l *Log) *Middleware {
	return &Middleware{provider: p, log: l, name: name}
}

func (m *Middleware) Get(ctx context.Context, key string) (string, error) {
	start := time.Now()
	v, err := m.provider.Get(ctx, key)
	m.record(OpGet, key, err, time.Since(start))
	return v, err
}

func (m *Middleware) Put(ctx context.Context, key, value string) error {
	start := time.Now()
	err := m.provider.Put(ctx, key, value)
	m.record(OpPut, key, err, time.Since(start))
	return err
}

func (m *Middleware) Delete(ctx context.Context, key string) error {
	start := time.Now()
	err := m.provider.Delete(ctx, key)
	m.record(OpDelete, key, err, time.Since(start))
	return err
}

func (m *Middleware) List(ctx context.Context) ([]string, error) {
	start := time.Now()
	keys, err := m.provider.List(ctx)
	m.record(OpList, "", err, time.Since(start))
	return keys, err
}

func (m *Middleware) record(op Operation, key string, err error, d time.Duration) {
	e := Entry{
		Provider:  m.name,
		Key:       key,
		Operation: op,
		Success:   err == nil,
		LatencyMs: d.Milliseconds(),
	}
	if err != nil {
		e.Error = err.Error()
	}
	m.log.Record(e)
}
