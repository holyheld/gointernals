package passwordutil

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func GeneratePasswordHash(pwd string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("can't generate password hash: %w", err)
	}

	return string(hash), nil
}

func ComparePasswordAndHash(pwd string, hashedPwd string) (bool, error) {
	byteHash := []byte(hashedPwd)

	err := bcrypt.CompareHashAndPassword(byteHash, []byte(pwd))
	if err != nil {
		return false, fmt.Errorf("can't compare password and hash: %w", err)
	}

	return true, nil
}
