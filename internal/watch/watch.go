// Package watch provides secret drift detection by periodically comparing
// live secret values against a known baseline snapshot.
package watch

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/vaultshift/internal/diff"
	"github.com/vaultshift/internal/provider"
)

// AlertFunc is called when drift is detected for a secret key.
type AlertFunc func(key string, result diff.Result)

// Watcher polls a provider and fires alerts on drift.
type Watcher struct {
	mu       sync.Mutex
	reg      *provider.Registry
	baseline map[string]string
	interval time.Duration
	alert    AlertFunc
}

// New creates a Watcher with the given poll interval and alert callback.
func New(reg *provider.Registry, interval time.Duration, alert AlertFunc) *Watcher {
	return &Watcher{
		reg:      reg,
		baseline: make(map[string]string),
		interval: interval,
		alert:    alert,
	}
}

// Snapshot records the current values for the given keys from providerName
// as the baseline to compare against on subsequent polls.
func (w *Watcher) Snapshot(ctx context.Context, providerName string, keys []string) error {
	p, err := w.reg.Get(providerName)
	if err != nil {
		return fmt.Errorf("watch: provider %q not found: %w", providerName, err)
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, k := range keys {
		val, err := p.GetSecret(ctx, k)
		if err != nil {
			return fmt.Errorf("watch: snapshot key %q: %w", k, err)
		}
		w.baseline[k] = val
	}
	return nil
}

// Start begins polling providerName for the given keys until ctx is cancelled.
func (w *Watcher) Start(ctx context.Context, providerName string, keys []string) error {
	p, err := w.reg.Get(providerName)
	if err != nil {
		return fmt.Errorf("watch: provider %q not found: %w", providerName, err)
	}
	go func() {
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				w.poll(ctx, p, keys.Lock()
	base := make(map[string]string, len(w.baseline))
	for k, v := range w.baseline {
		base[k] = v
	}
	w.mu.Unlock()

	current := make(map[string]string, len(keys))
	for _, k := range keys {
		val, err := p.GetSecret(ctx, k)
		if err == nil {
			current[k] = val
		}
	}
	results := diff.Compare(base, current)
	for _, r := range results {
		if r.Status != diff.Unchanged {
			w.alert(r.Key, r)
		}
	}
}
