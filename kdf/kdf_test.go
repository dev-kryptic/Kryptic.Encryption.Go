package kdf

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestArgon2idV1MatchesInteropVector(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "interop-vectors", "argon2id.json"))
	if err != nil {
		t.Fatal(err)
	}
	var vector struct {
		ParameterSetVersion int    `json:"parameterSetVersion"`
		Passphrase          string `json:"passphrase"`
		SaltHex             string `json:"saltHex"`
		DerivedKeyHex       string `json:"derivedKeyHex"`
	}
	if err := json.Unmarshal(data, &vector); err != nil {
		t.Fatal(err)
	}

	salt, _ := hex.DecodeString(vector.SaltHex)
	key, err := ForVersion(vector.ParameterSetVersion, vector.Passphrase, salt)
	if err != nil {
		t.Fatal(err)
	}
	if got := hex.EncodeToString(key); got != vector.DerivedKeyHex {
		t.Fatalf("derived key = %s, want %s", got, vector.DerivedKeyHex)
	}
}

func TestForVersionRejectsUnknownVersions(t *testing.T) {
	if _, err := ForVersion(2, "secret", make([]byte, SaltSize)); err == nil {
		t.Fatal("expected an error for an unknown parameter set version")
	}
}

func TestArgon2idV1RejectsBadSalt(t *testing.T) {
	if _, err := Argon2idV1("secret", make([]byte, 8)); err == nil {
		t.Fatal("expected an error for a short salt")
	}
}

func TestGenerateSalt(t *testing.T) {
	salt, err := GenerateSalt()
	if err != nil {
		t.Fatal(err)
	}
	if len(salt) != SaltSize {
		t.Fatalf("salt length = %d", len(salt))
	}
}
