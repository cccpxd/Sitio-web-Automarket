package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const (
	sessionTokenBytes = 32
	userIDBytes       = 16
)

// TokenIssuer implements the service.TokenIssuer port. Raw tokens are only
// returned to the caller; persistence uses the SHA-256 fingerprint.
type TokenIssuer struct{}

func NewTokenIssuer() *TokenIssuer {
	return &TokenIssuer{}
}

func (TokenIssuer) Issue() (string, error) {
	return randomHex(sessionTokenBytes)
}

func (TokenIssuer) Fingerprint(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}

// IDGenerator implements the service.IDGenerator port.
type IDGenerator struct{}

func NewIDGenerator() *IDGenerator {
	return &IDGenerator{}
}

func (IDGenerator) New() (string, error) {
	return randomHex(userIDBytes)
}

func randomHex(size int) (string, error) {
	buffer := make([]byte, size)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}
	return hex.EncodeToString(buffer), nil
}
