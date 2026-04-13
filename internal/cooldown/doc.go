// Package cooldown provides a key-scoped rate gate that enforces a minimum
// quiet period between successive operations on the same secret key.
//
// Use cooldown to prevent rapid re-rotation or re-propagation of secrets:
//
//	cd, err := cooldown.New(5 * time.Minute)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	if err := cd.Allow("prod/db/password"); err != nil {
//		// rotation was attempted too recently
//		fmt.Println("retry in", cd.Remaining("prod/db/password"))
//		return
//	}
//	// proceed with rotation ...
package cooldown
