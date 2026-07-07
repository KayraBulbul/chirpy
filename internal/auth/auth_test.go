package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWTHappyPath(t *testing.T) {
	randomUUID, err := uuid.NewRandom()
	if err != nil {
		t.Errorf("error creating randomUUID: %v", err)
	}

	jwtToken, err := MakeJWT(randomUUID, "hello", 5*time.Second)
	if err != nil {
		t.Errorf("error creating jwtToken: %v", err)
	}
	validatedUUID, err := ValidateJWT(jwtToken, "hello")
	if err != nil {
		t.Errorf("error validating jwtToken: %v", err)
	}
	if randomUUID != validatedUUID {
		t.Error("uuid does not match validatedUUID")
	}
}

func TestJWTWrongSecret(t *testing.T) {
	randomUUID, err := uuid.NewRandom()
	if err != nil {
		t.Errorf("error creating randomUUID: %v", err)
	}

	jwtToken, err := MakeJWT(randomUUID, "hello", 5*time.Second)
	if err != nil {
		t.Errorf("error creating jwtToken: %v", err)
	}
	_, err = ValidateJWT(jwtToken, "olleh")
	if err == nil {
		t.Error("was expecting a validation error but got nothing")
	}
}

func TestJWTExpiredToken(t *testing.T) {
	randomUUID, err := uuid.NewRandom()
	if err != nil {
		t.Errorf("error creating randomUUID: %v", err)
	}

	jwtToken, err := MakeJWT(randomUUID, "hello", -1*time.Second)
	if err != nil {
		t.Errorf("error creating jwtToken: %v", err)
	}
	_, err = ValidateJWT(jwtToken, "hello")
	if err == nil {
		t.Error("was expecting a validation error but got nothing")
	}
}
