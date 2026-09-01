package auth

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/google/uuid"
)

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func newJTI() string { return uuid.NewString() }
