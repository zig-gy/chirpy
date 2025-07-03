package auth

import "testing"

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