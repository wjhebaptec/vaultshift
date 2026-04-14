// Package cursor implements opaque pagination cursors for resumable secret
// list operations across cloud providers.
//
// A Cursor captures the provider name, current offset, optional key prefix,
// and page size. It is serialised to a URL-safe base64 token so callers can
// pass it between requests without exposing internal state.
//
// Usage:
//
//	m := cursor.New(cursor.WithPageSize(50))
//
//	// Start a new page
//	c := cursor.Cursor{Provider: "aws", Offset: 0, PageSize: 50}
//	token, _ := m.Encode(c)
//
//	// On the next request, decode and advance
//	c, _ = m.Decode(token)
//	nextCursor := m.Next(c)
package cursor
