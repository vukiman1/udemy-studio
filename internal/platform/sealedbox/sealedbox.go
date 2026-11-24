// Package sealedbox encrypts small secrets for storage with a key kept in a separate file, so a copy of the database alone reveals nothing.
package sealedbox

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

const (
	keyBytes    = 32
	keyFileMode = 0o600
)

var ErrUnreadable = errors.New("sealed value cannot be opened with this key")

type Box struct {
	aead cipher.AEAD
}

// LoadOrCreate reads the key file, creating it with a random key on first use.
func LoadOrCreate(path string) (*Box, error) {
	key, err := os.ReadFile(path) //nolint:gosec // reason: path is derived from the operator's --data-dir flag, never from a request
	if errors.Is(err, fs.ErrNotExist) {
		key = make([]byte, keyBytes)
		if _, err := rand.Read(key); err != nil {
			return nil, fmt.Errorf("generate encryption key: %w", err)
		}
		if err := os.WriteFile(path, key, keyFileMode); err != nil {
			return nil, fmt.Errorf("store encryption key %s: %w", path, err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("read encryption key %s: %w", path, err)
	}
	if len(key) != keyBytes {
		return nil, fmt.Errorf("encryption key %s has %d bytes, want %d", path, len(key), keyBytes)
	}
	return New(key)
}

func New(key []byte) (*Box, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}
	return &Box{aead: aead}, nil
}

// Seal binds the value to its label, so a sealed value moved under another label no longer opens.
func (box *Box) Seal(plaintext []byte, label string) (string, error) {
	nonce := make([]byte, box.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}
	return base64.StdEncoding.EncodeToString(box.aead.Seal(nonce, nonce, plaintext, []byte(label))), nil
}

func (box *Box) Open(sealed, label string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(sealed)
	if err != nil || len(data) < box.aead.NonceSize() {
		return nil, fmt.Errorf("%w: %q is not a sealed value", ErrUnreadable, label)
	}
	nonce, ciphertext := data[:box.aead.NonceSize()], data[box.aead.NonceSize():]
	plaintext, err := box.aead.Open(nil, nonce, ciphertext, []byte(label))
	if err != nil {
		return nil, fmt.Errorf("%w: %q", ErrUnreadable, label)
	}
	return plaintext, nil
}
