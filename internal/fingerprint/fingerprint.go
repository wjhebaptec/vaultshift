// Package fingerprint provides deterministic hashing of secret values
// to detect changes without exposing raw secret content.
package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// Fingerprinter computes and compares fingerprints of secret values.
type Fingerprinter struct {
	prefix string
}

// Option configures a Fingerprinter.
type Option func(*Fingerprinter)

// WithPrefix sets a namespace prefix mixed into every hash to prevent
// cross-environment collisions.
func WithPrefix(p string) Option {
	return func(f *Fingerprinter) {
		f.prefix = p
	}
}

// New returns a new Fingerprinter with the supplied options.
func New(opts ...Option) *Fingerprinter {
	f := &Fingerprinter{}
	for _, o := range opts {
		o(f)
	}
	return f
}

// Hash returns the SHA-256 hex digest of the given value, optionally
// namespaced by the configured prefix.
func (f *Fingerprinter) Hash(value string) string {
	h := sha256.New()
	if f.prefix != "" {
		h.Write([]byte(f.prefix + ":"))
	}
	h.Write([]byte(value))
	return hex.EncodeToString(h.Sum(nil))
}

// HashMap returns a map of key → fingerprint for every entry in secrets.
func (f *Fingerprinter) HashMap(secrets map[string]string) map[string]string {
	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		out[k] = f.Hash(v)
	}
	return out
}

// Changed returns the keys whose fingerprint differs between prev and next.
func Changed(prev, next map[string]string) []string {
	seen := make(map[string]struct{})
	var changed []string

	for k, nv := range next {
		seen[k] = struct{}{}
		if pv, ok := prev[k]; !ok || pv != nv {
			changed = append(changed, k)
		}
	}
	for k := range prev {
		if _, ok := seen[k]; !ok {
			changed = append(changed, k)
		}
	}
	sort.Strings(changed)
	return changed
}

// Summarise returns a single digest representing the combined state of all
// fingerprints in fp, useful for quick equality checks across full snapshots.
func Summarise(fp map[string]string) string {
	keys := make([]string, 0, len(fp))
	for k := range fp {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(fmt.Sprintf("%s=%s;", k, fp[k]))
	}
	h := sha256.Sum256([]byte(sb.String()))
	return hex.EncodeToString(h[:])
}
