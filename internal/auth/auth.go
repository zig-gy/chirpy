package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer: "chirpy",
		IssuedAt: jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject: userID.String(),
	})

	signedToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", fmt.Errorf("error signing token: %v", err)
	}
	return signedToken, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return uuid.Nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("error parsing token: %v", err)
	}

	userID, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, fmt.Errorf("error getting token subject: %v", err)
	}

	userUUID, err := uuid.Parse(userID)	
	if err != nil {
		return uuid.Nil, fmt.Errorf("error parsing uuid: %v", err)
	}
	return userUUID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	token := headers.Get("Authorization")
	if token == "" {
		return "", fmt.Errorf("token not found")
	}
	return token, nil
}