// Package envelope encrypts and opens Kryptic secret envelopes: the versioned
// container every Kryptic ciphertext travels in
// ("v1.<keyId>.<nonce>.<ciphertext+tag>", base64url without padding, AES-256-GCM).
// Secret values are end-to-end encrypted under the org key with associated data
// binding the ciphertext to its secret definition + environment, so rows cannot
// be swapped undetected.
package envelope

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

const (
	formatVersion = "v1"
	keySize       = 32
	nonceSize     = 12
	tagSize       = 16
)

// Envelope is a parsed Kryptic secret envelope.
type Envelope struct {
	KeyID             string
	Nonce             []byte
	CiphertextWithTag []byte
}

// Seal encrypts plaintext under a 256-bit key and returns the serialized envelope.
// associatedData is authenticated but not encrypted (GCM AAD).
func Seal(key []byte, keyID string, plaintext, associatedData []byte) (string, error) {
	if len(key) != keySize {
		return "", errors.New("key must be 32 bytes (AES-256)")
	}
	if !isValidKeyID(keyID) {
		return "", errors.New("key id must be non-empty and contain only [a-zA-Z0-9_-]")
	}

	nonce := make([]byte, nonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nil, nonce, plaintext, associatedData)

	return Envelope{KeyID: keyID, Nonce: nonce, CiphertextWithTag: ciphertext}.Serialize(), nil
}

// Parse parses the canonical serialized form.
func Parse(value string) (Envelope, error) {
	parts := strings.Split(value, ".")
	if len(parts) != 4 || parts[0] != formatVersion || !isValidKeyID(parts[1]) {
		return Envelope{}, errors.New("value is not a valid kryptic secret envelope")
	}

	enc := base64.RawURLEncoding
	nonce, err := enc.DecodeString(parts[2])
	if err != nil {
		return Envelope{}, fmt.Errorf("invalid envelope nonce encoding: %w", err)
	}
	ciphertext, err := enc.DecodeString(parts[3])
	if err != nil {
		return Envelope{}, fmt.Errorf("invalid envelope ciphertext encoding: %w", err)
	}
	if len(nonce) != nonceSize {
		return Envelope{}, errors.New("envelope nonce must be 12 bytes")
	}
	if len(ciphertext) < tagSize {
		return Envelope{}, errors.New("envelope ciphertext shorter than the authentication tag")
	}

	return Envelope{KeyID: parts[1], Nonce: nonce, CiphertextWithTag: ciphertext}, nil
}

// Serialize renders the canonical wire form.
func (e Envelope) Serialize() string {
	enc := base64.RawURLEncoding
	return strings.Join([]string{
		formatVersion, e.KeyID, enc.EncodeToString(e.Nonce), enc.EncodeToString(e.CiphertextWithTag),
	}, ".")
}

// Open decrypts a serialized envelope with the given 256-bit key. The
// associated data must match what the encrypting client bound the ciphertext
// to (for secret values: "secret:<definitionId>:env:<environmentId>").
func Open(key []byte, serialized string, associatedData []byte) ([]byte, error) {
	if len(key) != keySize {
		return nil, errors.New("key must be 32 bytes (AES-256)")
	}
	parsed, err := Parse(serialized)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := gcm.Open(nil, parsed.Nonce, parsed.CiphertextWithTag, associatedData)
	if err != nil {
		return nil, errors.New("envelope authentication failed - wrong key or tampered ciphertext")
	}
	return plaintext, nil
}

// SecretContext builds the associated data binding a secret value envelope to
// its row, matching the browser and the C# engine byte for byte.
func SecretContext(definitionID, environmentID string) []byte {
	return []byte("secret:" + strings.ToLower(definitionID) + ":env:" + strings.ToLower(environmentID))
}

func isValidKeyID(keyID string) bool {
	if keyID == "" || len(keyID) > 64 {
		return false
	}
	for _, c := range keyID {
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '_', c == '-':
		default:
			return false
		}
	}
	return true
}
