// Package envelope implements envelope encryption for vaultshift secrets.
//
// Envelope encryption separates the key used to encrypt data (DEK – data
// encryption key) from the key used to protect that DEK (KEK – key encryption
// key).  Only the encrypted DEK and the resulting ciphertext are persisted;
// the KEK is held exclusively in memory (or a dedicated KMS).
//
// Basic usage:
//
//	m := envelope.New()
//
//	// Register a KEK backed by AES-256-GCM.
//	key := make([]byte, 32) // load from KMS in production
//	enc, _ := envelope.NewAESEncryptor(key)
//	_ = m.RegisterKEK("primary", enc)
//
//	// Seal a secret.
//	sealed, _ := m.Seal("primary", []byte("s3cr3t"))
//
//	// Open it later.
//	plaintext, _ := m.Open(sealed)
//
Custom KEK implementations only need to satisfy the Encryptor interface,
making it straightforward to plug in HSM or cloud-KMS-backed encryptors.
package envelope
