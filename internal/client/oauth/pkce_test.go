package oauth

import (
	"regexp"
	"testing"
)

var base64URL = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)

func TestS256MatchesTheRFC7636Example(t *testing.T) {
	got := S256("dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk")
	if want := "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"; got != want {
		t.Errorf("S256 = %q, want %q", got, want)
	}
}

func TestNewPKCEIsRandomAndConsistent(t *testing.T) {
	first, err := NewPKCE()
	if err != nil {
		t.Fatalf("NewPKCE: %v", err)
	}
	second, err := NewPKCE()
	if err != nil {
		t.Fatalf("NewPKCE: %v", err)
	}
	if !base64URL.MatchString(first.Verifier) || first.Verifier == second.Verifier {
		t.Errorf("verifiers %q and %q, want two distinct 43 character base64url strings", first.Verifier, second.Verifier)
	}
	if first.Challenge != S256(first.Verifier) {
		t.Errorf("challenge %q does not match the verifier", first.Challenge)
	}
}

func TestRandomTokenIsBase64URL(t *testing.T) {
	token, err := RandomToken()
	if err != nil || !base64URL.MatchString(token) {
		t.Errorf("RandomToken = %q, %v, want 43 base64url characters", token, err)
	}
}
