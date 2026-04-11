// Package encrypt provides AES-GCM symmetric encryption and decryption
// for vaultshift secret values.
//
// Usage:
//
//	e, err := encrypt.New([]byte("32-byte-secret-key-here!!!!!!!!"))
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	ciphertext, err := e.Encrypt("my-secret-value")
//	plaintext, err := e.Decrypt(ciphertext)
//
// Keys must be exactly 16, 24, or 32 bytes corresponding to AES-128,
// AES-192, and AES-256 respectively. Each call to Encrypt uses a
// randomly generated nonce, making the output non-deterministic.
package encrypt
