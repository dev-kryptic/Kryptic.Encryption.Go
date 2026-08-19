// Package kdf derives keys from low-entropy secrets with Argon2id, matching
// the C# Argon2Parameters byte for byte (locked by interop-vectors/argon2id.json).
// The CI path and Kubernetes operator use it to turn a machine's client secret
// into the key that unwraps the machine private key.
package kdf

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/argon2"
)

const (
	// SaltSize matches Argon2KeyDerivation.SaltSizeBytes.
	SaltSize = 16
	// KeySize matches Argon2KeyDerivation.DerivedKeySizeBytes.
	KeySize = 32
)

// GenerateSalt returns a fresh 16-byte salt.
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, SaltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	return salt, nil
}

// Argon2idV1 is parameter set 1: 64 MiB, 3 passes, 4 lanes, 32-byte output.
func Argon2idV1(secret string, salt []byte) ([]byte, error) {
	if len(salt) != SaltSize {
		return nil, fmt.Errorf("argon2 salt must be %d bytes", SaltSize)
	}
	return argon2.IDKey([]byte(secret), salt, 3, 64*1024, 4, KeySize), nil
}

// ForVersion dispatches on the parameter set version stored with the record.
func ForVersion(version int, secret string, salt []byte) ([]byte, error) {
	if version != 1 {
		return nil, fmt.Errorf("unknown Argon2 parameter set version %d", version)
	}
	return Argon2idV1(secret, salt)
}
