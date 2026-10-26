package config

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/markupp-labs/markupp-cli/internal/failure"
)

func TestSetGetUnsetEveryKey(t *testing.T) {
	values := map[string]string{KeyURL: "https://markupp.example.com", KeyTenant: "t1", KeyVault: "v1"}
	var settings Settings
	for _, key := range Keys {
		if err := settings.Set(key, values[key]); err != nil {
			t.Fatalf("Set(%q): %v", key, err)
		}
		got, err := settings.Get(key)
		if err != nil || got != values[key] {
			t.Fatalf("Get(%q) = %q, %v, want %q", key, got, err, values[key])
		}
		if err := settings.Unset(key); err != nil {
			t.Fatalf("Unset(%q): %v", key, err)
		}
		if got, _ := settings.Get(key); got != "" {
			t.Errorf("after Unset(%q), Get = %q, want empty", key, got)
		}
	}
}

func TestUnknownKeyIsAUsageErrorNamingTheKey(t *testing.T) {
	var settings Settings
	checks := map[string]error{
		"Set":   settings.Set("token", "x"),
		"Unset": settings.Unset("token"),
	}
	_, checks["Get"] = settings.Get("token")
	for name, err := range checks {
		if failure.KindOf(err) != failure.Usage || !strings.Contains(fmt.Sprint(err), `"token"`) {
			t.Errorf("%s(token) error = %v, want a usage error naming the key", name, err)
		}
	}
}

func TestSetRejectsEmptyValue(t *testing.T) {
	var settings Settings
	if err := settings.Set(KeyVault, "  "); failure.KindOf(err) != failure.Usage {
		t.Fatalf("Set(vault, blank) error = %v, want usage", err)
	}
}

func TestNormalizeURL(t *testing.T) {
	cases := map[string]string{
		"https://markupp.example.com/": "https://markupp.example.com",
		"https://example.com/markupp/": "https://example.com/markupp",
	}
	for in, want := range cases {
		got, err := NormalizeURL(in)
		if err != nil || got != want {
			t.Errorf("NormalizeURL(%q) = %q, %v, want %q", in, got, err, want)
		}
	}
}

func TestNormalizeURLTrimsSpaces(t *testing.T) {
	got, err := NormalizeURL("  http://localhost:8080/ ")
	if err != nil || got != "http://localhost:8080" {
		t.Errorf("NormalizeURL with spaces = %q, %v, want http://localhost:8080", got, err)
	}
}

func TestNormalizeURLRejectsNonHTTP(t *testing.T) {
	for _, in := range []string{"", "markupp.example.com", "ftp://x", "https://", "https://x?q=1"} {
		_, err := NormalizeURL(in)
		if failure.KindOf(err) != failure.Usage || !strings.Contains(fmt.Sprint(err), fmt.Sprintf("%q", in)) {
			t.Errorf("NormalizeURL(%q) error = %v, want a usage error naming the value", in, err)
		}
	}
}

func TestCredentialStringNeverShowsTokens(t *testing.T) {
	cred := Credential{AccessToken: "secret-access", RefreshToken: "secret-refresh", ExpiresAt: time.Unix(0, 0)}
	for _, text := range []string{cred.String(), fmt.Sprintf("%v", cred), fmt.Sprintf("%+v", cred)} {
		if strings.Contains(text, "secret") {
			t.Errorf("credential rendered as %q, which leaks a token", text)
		}
	}
}
