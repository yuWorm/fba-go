package secret

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

const (
	DefaultBytes = 32
	MinimumBytes = 32
	MaximumBytes = 1024
)

// Generate returns URL-safe, unpadded key material backed by the requested
// number of bytes from the operating system's cryptographic random source.
func Generate(size int) (string, error) {
	if size < MinimumBytes || size > MaximumBytes {
		return "", fmt.Errorf("secret size must be between %d and %d bytes", MinimumBytes, MaximumBytes)
	}
	raw := make([]byte, size)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("read cryptographic randomness: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
