package oauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// PKCE is the verifier kept by mu and the S256 challenge sent to the authorize
// endpoint (RFC 7636).
type PKCE struct {
	Verifier  string
	Challenge string
}

// NewPKCE returns a random verifier and its S256 challenge.
func NewPKCE() (PKCE, error) {
	verifier, err := RandomToken()
	if err != nil {
		return PKCE{}, err
	}
	return PKCE{Verifier: verifier, Challenge: S256(verifier)}, nil
}

// RandomToken returns 32 bytes from crypto/rand in base64url, 43 characters: a PKCE
// verifier or an OAuth state.
func RandomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("reading 32 random bytes for the OAuth flow: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// S256 returns the challenge of verifier: base64url(sha256(verifier)).
func S256(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
