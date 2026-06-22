package auth

import (
	"fmt"

	"github.com/alexedwards/argon2id"
)

func HashPassword(password string) (string, error) {

	// what else am I supposed to pass
	hashed_passwords, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", fmt.Errorf("failed to hash the password: %v", err)
	}

	return hashed_passwords, nil
}
