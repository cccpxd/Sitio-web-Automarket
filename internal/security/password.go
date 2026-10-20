package security

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	// DefaultIterations follows the OWASP recommendation for PBKDF2-HMAC-SHA256.
	DefaultIterations = 600_000

	passwordAlgorithm = "pbkdf2-sha256"
	saltLength        = 16
	keyLength         = 32
)

// DummyPasswordHash is verified for unknown accounts to reduce timing-based
// user enumeration. It is not a secret and does not belong to any account.
const DummyPasswordHash = "pbkdf2-sha256$600000$YXV0b21hcmtldC1kdW1teQ$F5F0RmYR6EvdmWFjLJ4QSj4qTl9AkXq9URcXHZQA0SE"

// PBKDF2Hasher implements the PasswordHasher port with PBKDF2-HMAC-SHA256.
type PBKDF2Hasher struct {
	iterations int
}

func NewPBKDF2Hasher(iterations int) (*PBKDF2Hasher, error) {
	if iterations <= 0 {
		return nil, errors.New("password hasher: iterations must be positive")
	}
	return &PBKDF2Hasher{iterations: iterations}, nil
}

func (h *PBKDF2Hasher) Hash(password string) (string, error) {
	if password == "" {
		return "", errors.New("password hasher: password is empty")
	}

	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("password hasher: generate salt: %w", err)
	}

	key, err := pbkdf2.Key(sha256.New, password, salt, h.iterations, keyLength)
	if err != nil {
		return "", fmt.Errorf("password hasher: derive key: %w", err)
	}

	return fmt.Sprintf(
		"%s$%d$%s$%s",
		passwordAlgorithm,
		h.iterations,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

func (h *PBKDF2Hasher) Verify(password, encodedHash string) bool {
	algorithm, iterations, salt, expected, ok := parsePasswordHash(encodedHash)
	if !ok || algorithm != passwordAlgorithm || password == "" {
		return false
	}

	candidate, err := pbkdf2.Key(sha256.New, password, salt, iterations, len(expected))
	if err != nil {
		return false
	}

	return subtle.ConstantTimeCompare(candidate, expected) == 1
}

func parsePasswordHash(encoded string) (algorithm string, iterations int, salt, key []byte, ok bool) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 {
		return "", 0, nil, nil, false
	}

	iterations, err := strconv.Atoi(parts[1])
	if err != nil || iterations <= 0 {
		return "", 0, nil, nil, false
	}

	salt, err = base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil || len(salt) < saltLength {
		return "", 0, nil, nil, false
	}

	key, err = base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil || len(key) == 0 {
		return "", 0, nil, nil, false
	}

	return parts[0], iterations, salt, key, true
}
