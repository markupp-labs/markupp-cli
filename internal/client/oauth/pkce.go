package oauth

// PKCE is the verifier kept by mu and the S256 challenge sent to the authorize endpoint.
type PKCE struct {
	Verifier  string
	Challenge string
}

// NewPKCE returns a random verifier and its challenge.
func NewPKCE() (PKCE, error) { return PKCE{}, errNotImplemented }

// RandomToken returns 32 random bytes in base64url.
func RandomToken() (string, error) { return "", errNotImplemented }

// S256 returns the challenge of verifier.
func S256(string) string { return "" }
