package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCreateHash(t *testing.T) {
	pass := "Hola"
	_, err := HashPassword(pass)
	if err != nil {
		t.Errorf("error craeting hash: %v", err)
	}
}

func TestCheckPasswordHash(t *testing.T) {
	pass := "hola"
	hash, err := HashPassword(pass)
	if err != nil {
		t.Errorf("error creating hash:: %v", err)
	}
	err = CheckPasswordHash(pass, hash)
	if err != nil {
		t.Errorf("error checking password hash: %v", err)
	}
}

func TestCreateAndValidateJWT (t *testing.T) {
	id := uuid.New()
	time, err := time.ParseDuration("10h")
	if err != nil {
		t.Errorf("Error parsing duration: %v\n", err)
		return
	}
	tokenSecret := "hola"

	jwt, err := MakeJWT(id, tokenSecret, time)
	if err != nil {
		t.Errorf("Error making jwt: %v\n", err)
		return
	}

	newID, err := ValidateJWT(jwt, tokenSecret)
	if err != nil {
		t.Errorf("Error validating jwt: %v\n", err)
		return
	}
	
	if newID != id {
		t.Errorf("UUID doesn't coincide")
		return
	}
}