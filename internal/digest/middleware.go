package digest

import (
	"context"
	"fmt"

	"github.com/vaultshift/internal/provider"
)

// VerifyingProvider wraps a provider.Provider and verifies digests on Get.
type VerifyingProvider struct {
	inner  provider.Provider
	signer *Signer
}

// Wrap returns a VerifyingProvider that signs on Put and verifies on Get.
func Wrap(inner provider.Provider, signer *Signer) (*VerifyingProvider, error) {
	if inner == nil {
		return nil, fmt.Errorf("digest: inner provider must not be nil")
	}
	if signer == nil {
		return nil, fmt.Errorf("digest: signer must not be nil")
	}
	return &VerifyingProvider{inner: inner, signer: signer}, nil
}

// Put stores the secret and records its digest.
func (v *VerifyingProvider) Put(ctx context.Context, key, value string) error {
	if err := v.inner.Put(ctx, key, value); err != nil {
		return err
	}
	return v.signer.Sign(key, value)
}

// Get retrieves the secret and verifies its digest.
func (v *VerifyingProvider) Get(ctx context.Context, key string) (string, error) {
	val, err := v.inner.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if verr := v.signer.Verify(key, val); verr != nil {
		return "", fmt.Errorf("digest: tamper detected for %q: %w", key, verr)
	}
	return val, nil
}

// Delete removes the secret and its digest.
func (v *VerifyingProvider) Delete(ctx context.Context, key string) error {
	if err := v.inner.Delete(ctx, key); err != nil {
		return err
	}
	v.signer.Delete(key)
	return nil
}

// List delegates to the inner provider.
func (v *VerifyingProvider) List(ctx context.Context) ([]string, error) {
	return v.inner.List(ctx)
}
