package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
)

func NewSessionToken() (string, error) {
	var token [32]byte
	if _, err := rand.Read(token[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(token[:]), nil
}

func SessionTokenEqual(expected, supplied string) bool {
	if len(expected) != len(supplied) || expected == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(supplied)) == 1
}

func ValidateSessionToken(token string) error {
	if len(token) != 64 {
		return errors.New("invalid session token")
	}
	if _, err := hex.DecodeString(token); err != nil {
		return errors.New("invalid session token")
	}
	return nil
}
