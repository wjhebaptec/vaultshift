package pipeline

import (
	"context"
	"fmt"
)

// Builder provides a fluent API for constructing common vaultshift pipelines.
type Builder struct {
	p *Pipeline
}

// NewBuilder returns a Builder wrapping a fresh Pipeline.
func NewBuilder() *Builder {
	return &Builder{p: New()}
}

// WithStep is a generic escape-hatch to add any Step.
func (b *Builder) WithStep(s Step) *Builder {
	b.p.Add(s)
	return b
}

// WithValidation adds a step that validates the payload value using the
// provided predicate. The step fails if the predicate returns false.
func (b *Builder) WithValidation(name string, pred func(value string) bool) *Builder {
	b.p.Add(Step{
		Name: fmt.Sprintf("validate:%s", name),
		Run: func(_ context.Context, pl *Payload) error {
			if !pred(pl.Value) {
				return fmt.Errorf("validation %q failed for key %q", name, pl.Key)
			}
			return nil
		},
	})
	return b
}

// WithTransform adds a step that transforms the payload value in-place.
func (b *Builder) WithTransform(name string, fn func(value string) (string, error)) *Builder {
	b.p.Add(Step{
		Name: fmt.Sprintf("transform:%s", name),
		Run: func(_ context.Context, pl *Payload) error {
			v, err := fn(pl.Value)
			if err != nil {
				return fmt.Errorf("transform %q: %w", name, err)
			}
			pl.Value = v
			return nil
		},
	})
	return b
}

// WithMetaTag adds a step that sets a metadata key on the payload.
func (b *Builder) WithMetaTag(key, value string) *Builder {
	b.p.Add(Step{
		Name: fmt.Sprintf("meta:%s", key),
		Run: func(_ context.Context, pl *Payload) error {
			if pl.Meta == nil {
				pl.Meta = map[string]string{}
			}
			pl.Meta[key] = value
			return nil
		},
	})
	return b
}

// Build returns the constructed Pipeline.
func (b *Builder) Build() *Pipeline { return b.p }
