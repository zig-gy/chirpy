package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (hashedPassword string, err error) {
	bytesPass := []byte(password)
	hashBytes, err := bcrypt.GenerateFromPassword(bytesPass, 1)
	if err != nil {
		return "", fmt.Errorf("error generating hash: %v", err)
	}
	hashedPassword = string(hashBytes)
	return
}

func CheckPasswordHash(password, hash string) error {
	bytesHash := []byte(hash)
	bytesPass := []byte(password)
	if err := bcrypt.CompareHashAndPassword(bytesHash, bytesPass); err != nil {
		return fmt.Errorf("error comparing hash and password: %v", err)
	}
	return nil
}