package hash

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"golang.org/x/crypto/argon2"
	"log"
)

const (
	saltLength = 16
	keyLength  = 32
)

// GenerateSalt creates a new random salt.
func GenerateSalt() []byte {
	salt := make([]byte, saltLength)

	_, err := rand.Read(salt)
	if err != nil {
		log.Panicf("Failed to generate salt: %v", err)
	}

	return salt
}

// HashPassword hashes the password using Argon2.
func HashPassword(password string, salt []byte) string {
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, keyLength)

	return base64.RawStdEncoding.EncodeToString(hash)
}

// comparePasswords checks if the provided password hashes to the same value.
func comparePasswords(password string, salt []byte, encodedHash string) bool {
	hash, err := base64.RawStdEncoding.DecodeString(encodedHash)
	if err != nil {
		return false
	}

	otherHash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, keyLength)

	return subtle.ConstantTimeCompare(hash, otherHash) == 1
}
