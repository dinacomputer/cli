package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

// PKCE holds a code verifier / challenge pair.
type PKCE struct {
	Verifier  string
	Challenge string
}

// NewPKCE generates a PKCE code verifier and S256 challenge.
func NewPKCE() (*PKCE, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	verifier := base64.RawURLEncoding.EncodeToString(buf)
	h := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(h[:])
	return &PKCE{Verifier: verifier, Challenge: challenge}, nil
}
