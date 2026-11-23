package library

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

const idHexLength = 12

func shortID(kind string, parts ...string) string {
	sum := sha256.Sum256([]byte(kind + "\x00" + strings.Join(parts, "\x00")))
	return hex.EncodeToString(sum[:])[:idHexLength]
}
