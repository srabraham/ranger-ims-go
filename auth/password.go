package auth

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
)

func VerifyPassword(password, storedValue string) (bool, error) {
	salt, storedHash, found := strings.Cut(storedValue, ":")
	if !found {
		return false, fmt.Errorf("invalid hashed password")
	}
	return hash(password, salt) == storedHash, nil
}

func hash(password, salt string) string {
	hasher := sha1.New()
	hasher.Write([]byte(salt + password))
	return hex.EncodeToString(hasher.Sum(nil))
}
