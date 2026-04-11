// Package event implements a lightweight publish-subscribe event bus used
// throughout vaultshift to decouple components that need to react to secret
// lifecycle changes.
//
// Usage:
//
//	bus := event.New()
//
//	bus.Subscribe(event.TypeRotated, func(e event.Event) {
//		fmt.Printf("rotated %s on %s\n", e.Key, e.Provider)
//	})
//
//	bus.Publish(event.Event{
//		Type:     event.TypeRotated,
//		Key:      "db/password",
//		Provider: "aws",
//	})
package event
