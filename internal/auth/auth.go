package auth

import (
	"errors"
	"net/http"
	"runtime"
	"strings"

	"github.com/alexedwards/argon2id"
)

func HashPassword(password string) (string, error) {
	params := &argon2id.Params{
		Memory:      128 * 1024,
		Iterations:  4,
		Parallelism: uint8(runtime.NumCPU()),
		SaltLength:  16,
		KeyLength:   32,
	}
	hashedPassword, err := argon2id.CreateHash(password, params)
	if err != nil {
		return "", err
	}
	return hashedPassword, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}
	return match, nil
}

func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("could not find authorization header")
	}

	apiKey, ok := strings.CutPrefix(authHeader, "ApiKey ")
	if !ok {
		return "", errors.New("malformed authorization header")
	}

	return apiKey, nil
}
