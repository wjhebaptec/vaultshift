// Package watermark provides utilities for embedding and verifying hidden
// markers inside secret values.
//
// A Marker stamps a lightweight tag onto the end of a secret value that
// encodes a namespace prefix and a caller-supplied label. This allows
// downstream systems to detect whether a value originated from a known
// rotation, environment, or pipeline stage.
//
// Example usage:
//
//	m, err := watermark.New("prod", watermark.WithSeparator(":"))
//	if err != nil { ... }
//
//	stamped := m.Stamp("rotation-2024", originalValue)
//
//	clean, label, err := m.Strip(stamped)
//	if err != nil { ... }
//
// The watermark tag is appended to the value and does not alter the
// meaningful content when stripped before use.
package watermark
