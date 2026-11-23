package chromecookie

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"errors"
	"testing"
)

const testPassword = "test-safe-storage-password" //nolint:gosec // reason: fixture password for round-trip tests

func encryptForTest(t *testing.T, value, hostKey string, schemaVersion int) []byte {
	t.Helper()
	key, err := deriveKey([]byte(testPassword))
	if err != nil {
		t.Fatalf("derive key: %v", err)
	}
	plaintext := []byte(value)
	if schemaVersion >= hostDigestSchemaVersion {
		digest := sha256.Sum256([]byte(hostKey))
		plaintext = append(digest[:], plaintext...)
	}
	padding := aes.BlockSize - len(plaintext)%aes.BlockSize
	plaintext = append(plaintext, bytes.Repeat([]byte{byte(padding)}, padding)...)
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatalf("cipher: %v", err)
	}
	ciphertext := make([]byte, len(plaintext))
	cipher.NewCBCEncrypter(block, bytes.Repeat([]byte{initializationVectorByte}, aes.BlockSize)).CryptBlocks(ciphertext, plaintext)
	return append([]byte(encryptedValuePrefix), ciphertext...)
}

func TestDecryptValueRoundTrips(t *testing.T) {
	key, _ := deriveKey([]byte(testPassword))
	for _, schemaVersion := range []int{23, 24} {
		encrypted := encryptForTest(t, "token-value", "acme.udemy.com", schemaVersion)
		got, err := decryptValue(encrypted, key, "acme.udemy.com", schemaVersion)
		if err != nil || string(got) != "token-value" {
			t.Errorf("schema %d: got %q, %v", schemaVersion, got, err)
		}
	}
}

func TestDecryptValueRejectsWhatItCannotTrust(t *testing.T) {
	key, _ := deriveKey([]byte(testPassword))
	valid := encryptForTest(t, "token-value", "acme.udemy.com", 24)
	wrongKey, _ := deriveKey([]byte("another password"))
	cases := []struct {
		name      string
		encrypted []byte
		key       []byte
		hostKey   string
		want      error
	}{
		{name: "app-bound v20 value", encrypted: append([]byte("v20"), valid[3:]...), key: key, hostKey: "acme.udemy.com", want: ErrUnsupportedEncryption},
		{name: "truncated ciphertext", encrypted: valid[:len(valid)-3], key: key, hostKey: "acme.udemy.com", want: ErrBadPadding},
		{name: "wrong keychain password", encrypted: valid, key: wrongKey, hostKey: "acme.udemy.com", want: nil},
		{name: "cookie copied from another host", encrypted: valid, key: key, hostKey: "evil.udemy.com", want: ErrHostDigestMismatch},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := decryptValue(testCase.encrypted, testCase.key, testCase.hostKey, 24)
			if testCase.want == nil {
				if err == nil && string(got) == "token-value" {
					t.Fatalf("decrypting with the wrong key returned the real value")
				}
				return
			}
			if !errors.Is(err, testCase.want) {
				t.Errorf("error = %v, want %v", err, testCase.want)
			}
		})
	}
}
