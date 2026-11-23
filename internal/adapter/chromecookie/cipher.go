package chromecookie

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/pbkdf2"
	"crypto/sha1" //nolint:gosec // reason: Chrome derives its cookie key with PBKDF2-SHA1, we only mirror it
	"crypto/sha256"
	"fmt"
)

const (
	encryptedValuePrefix     = "v10"
	pbkdf2Salt               = "saltysalt"
	pbkdf2Iterations         = 1003
	aesKeyBytes              = 16
	hostDigestSchemaVersion  = 24
	initializationVectorByte = ' '
)

func deriveKey(password []byte) ([]byte, error) {
	key, err := pbkdf2.Key(sha1.New, string(password), []byte(pbkdf2Salt), pbkdf2Iterations, aesKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("derive cookie key: %w", err)
	}
	return key, nil
}

func decryptValue(encrypted, key []byte, hostKey string, schemaVersion int) ([]byte, error) {
	if !bytes.HasPrefix(encrypted, []byte(encryptedValuePrefix)) {
		return nil, ErrUnsupportedEncryption
	}
	ciphertext := encrypted[len(encryptedValuePrefix):]
	if len(ciphertext) == 0 || len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("%w: ciphertext of %d bytes", ErrBadPadding, len(ciphertext))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cookie cipher: %w", err)
	}
	plaintext := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, bytes.Repeat([]byte{initializationVectorByte}, aes.BlockSize)).CryptBlocks(plaintext, ciphertext)
	plaintext, err = stripPKCS7(plaintext)
	if err != nil {
		return nil, err
	}
	if schemaVersion < hostDigestSchemaVersion {
		return plaintext, nil
	}
	hostDigest := sha256.Sum256([]byte(hostKey))
	if !bytes.HasPrefix(plaintext, hostDigest[:]) {
		return nil, ErrHostDigestMismatch
	}
	return plaintext[sha256.Size:], nil
}

func stripPKCS7(data []byte) ([]byte, error) {
	padding := int(data[len(data)-1])
	if padding == 0 || padding > aes.BlockSize || padding > len(data) {
		return nil, ErrBadPadding
	}
	paddingStart := len(data) - padding
	if !bytes.Equal(data[paddingStart:], bytes.Repeat([]byte{byte(padding)}, padding)) {
		return nil, ErrBadPadding
	}
	return data[:paddingStart], nil
}
