package envelope

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
	"testing"
)

func testKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	return key
}

func TestSealOpenRoundTrip(t *testing.T) {
	key := testKey(t)
	associatedData := SecretContext("A1B2C3D4-0000-0000-0000-000000000001", "a1b2c3d4-0000-0000-0000-000000000002")
	serialized, err := Seal(key, "org-key-1", []byte("postgres://user:pw@host/db"), associatedData)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	plaintext, err := Open(key, serialized, associatedData)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if string(plaintext) != "postgres://user:pw@host/db" {
		t.Fatalf("plaintext = %q", plaintext)
	}
}

func TestOpenRejectsWrongAssociatedData(t *testing.T) {
	key := testKey(t)
	right := SecretContext("def-1", "env-1")
	serialized, err := Seal(key, "org-key-1", []byte("value"), right)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := Open(key, serialized, SecretContext("def-1", "env-2")); err == nil {
		t.Fatal("expected authentication failure when the envelope is bound to another environment")
	}
}

func TestOpenRejectsWrongKey(t *testing.T) {
	right := testKey(t)
	serialized, err := Seal(right, "org-key-1", []byte("value"), nil)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := Open(testKey(t), serialized, nil); err == nil {
		t.Fatal("expected authentication failure with the wrong key")
	}
}

func TestParseRejectsMalformedEnvelopes(t *testing.T) {
	enc := base64.RawURLEncoding
	shortNonce := "v1.k." + enc.EncodeToString(make([]byte, 4)) + "." + enc.EncodeToString(make([]byte, 32))
	shortCiphertext := "v1.k." + enc.EncodeToString(make([]byte, nonceSize)) + "." + enc.EncodeToString(make([]byte, 8))

	for _, value := range []string{
		"", "v1", "v2.k.a.b", "v1.k.!!!.b", "not an envelope",
		shortNonce, shortCiphertext,
	} {
		if _, err := Parse(value); err == nil {
			t.Fatalf("Parse(%q) accepted a malformed envelope", value)
		}
	}
}

func TestSecretContextFormat(t *testing.T) {
	got := string(SecretContext("ABC-123", "DEF-456"))
	want := "secret:abc-123:env:def-456"
	if got != want {
		t.Fatalf("SecretContext = %q, want %q", got, want)
	}
}

func TestSealRejectsBadKeyID(t *testing.T) {
	if _, err := Seal(testKey(t), "bad key", []byte("x"), nil); err == nil {
		t.Fatal("expected an error for a key id with a space")
	}
}

func TestSerializeRoundTrip(t *testing.T) {
	key := testKey(t)
	serialized, err := Seal(key, "key_abc", []byte("x"), nil)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := Parse(serialized)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Serialize() != serialized {
		t.Fatal("serialize(parse(x)) != x")
	}
	if !strings.HasPrefix(serialized, "v1.key_abc.") {
		t.Fatalf("unexpected prefix: %s", serialized)
	}
}
