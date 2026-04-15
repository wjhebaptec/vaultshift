// Package signal implements a lightweight named-signal bus for vaultshift.
//
// Components can register handlers for named signals and emit those signals
// with arbitrary payloads, enabling decoupled communication between modules
// such as rotation, sync, and audit without direct dependencies.
//
// Example:
//
//	bus := signal.New()
//
//	_ = bus.On("secret.rotated", func(name string, payload any) {
//		fmt.Println("rotated:", payload)
//	})
//
//	bus.Emit("secret.rotated", "prod/db-password")
package signal
