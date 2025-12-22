package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"strings"

	"golang.org/x/crypto/argon2"
)

func HashPassword(password string) (string, error) {
	// Generate a random salt
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", errors.New("failed to generate salt")
	}

	// Hash the password using Argon2
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	// Encode salt and hash to base64
	saltBase64 := base64.StdEncoding.EncodeToString(salt)
	hashBase64 := base64.StdEncoding.EncodeToString(hash)

	// Return in format: salt.hash
	return saltBase64 + "." + hashBase64, nil
}

func VerifyPassword(inputPassword, storedPassword string) error {
	parts := strings.Split(storedPassword, ".")
	if len(parts) != 2 {
		return errors.New("invalid encoded has format")
	}

	saltBase64 := parts[0]
	hashedPasswordBase64 := parts[1]

	salt, err := base64.StdEncoding.DecodeString(saltBase64)
	if err != nil {
		return errors.New("failed to decode the salt")
	}

	hashedPassword, err := base64.StdEncoding.DecodeString(hashedPasswordBase64)
	if err != nil {
		return errors.New("failed to decode the hashed password")
	}

	hash := argon2.IDKey([]byte(inputPassword), salt, 1, 64*1024, 4, 32)
	if len(hash) != len(hashedPassword) {
		return errors.New("incorrect password")
	}

	if subtle.ConstantTimeCompare(hash, hashedPassword) != 1 {
		return errors.New("incorrect password")
	}

	return nil
}
