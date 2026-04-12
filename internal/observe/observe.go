// Package observe provides a lightweight observation layer that combines
// metrics recording, audit logging, and event publishing into a single
// coordinated hook for secret provider operations.
package observe

import (
	"context"
	"time"

	"github.com/vaultshift/internal/audit"
	"github.com/vaultshift/internal/event"
	"github.com/vaultshift/internal/metrics"
)

// Event kinds emitted by the observer.
const (
	KindGet    = "observe.get"
	KindPut    = "observe.put"
	KindDelete = "observe.delete"
)

// Observer coordinates metrics, audit, and event emission for a named provider.
type Observer struct {
	provider string
	collector *metrics.Collector
	logger    *audit.Logger
	bus       *event.Bus
}

// New creates an Observer for the given provider name.
func New(provider string, c *metrics.Collector, l *audit.Logger, b *event.Bus) *Observer {
	return &Observer{
		provider:  provider,
		collector: c,
		logger:    l,
		bus:       b,
	}
}

// Record captures a completed operation, recording a metric entry, writing an
// audit log line, and publishing an event on the bus.
func (o *Observer) Record(ctx context.Context, kind, key string, err error) {
	status := "ok"
	if err != nil {
		status = "error"
	}

	o.collector.Record(metrics.Entry{
		Type:      kind,
		Provider:  o.provider,
		Key:       key,
		Status:    status,
		Timestamp: time.Now(),
	})

	o.logger.Log(audit.Event{
		Operation: kind,
		Provider:  o.provider,
		Key:       key,
		Status:    status,
		Timestamp: time.Now(),
	})

	o.bus.Publish(event.Event{
		Kind:    kind,
		Payload: map[string]string{"provider": o.provider, "key": key, "status": status},
	})
}
