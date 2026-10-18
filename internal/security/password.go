package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"
)

const (
	passwordIterations = 20000
)

func HashPassword(password string) (hash string, salt string, err error) {
	if password == "" {
		return "", "", errors.New("la contraseña no puede estar vacía")
	}

	saltBytes := make([]byte, 16)
	if _, err = rand.Read(saltBytes); err != nil {
		return "", "", err
	}

	hashBytes := hashPasswordBytes([]byte(password), saltBytes)
	return hex.EncodeToString(hashBytes), hex.EncodeToString(saltBytes), nil
}

func VerifyPassword(password, saltHex, storedHashHex string) bool {
	if password == "" || saltHex == "" || storedHashHex == "" {
		return false
	}

	saltBytes, err := hex.DecodeString(saltHex)
	if err != nil {
		return false
	}

	storedHashBytes, err := hex.DecodeString(storedHashHex)
	if err != nil {
		return false
	}

	candidate := hashPasswordBytes([]byte(password), saltBytes)
	if len(candidate) != len(storedHashBytes) {
		return false
	}

	for i := range candidate {
		if candidate[i] != storedHashBytes[i] {
			return false
		}
	}
	return true
}

func hashPasswordBytes(password, salt []byte) []byte {
	value := append([]byte{}, password...)
	for i := 0; i < passwordIterations; i++ {
		hash := sha256.Sum256(append(value, salt...))
		value = hash[:]
	}
	return value
}

func GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func GenerateID() string {
	id, err := GenerateToken()
	if err != nil {
		return time.Now().UTC().Format(time.RFC3339Nano)
	}
	return id
}
