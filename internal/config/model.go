package config

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/markupp-labs/markupp-cli/internal/failure"
)

// Settable keys of the configuration file.
const (
	KeyURL    = "url"
	KeyTenant = "tenant"
	KeyVault  = "vault"
)

// Keys lists the keys mu config set and unset accept, in display order.
var Keys = []string{KeyURL, KeyTenant, KeyVault}

// Settings is the content of the configuration file.
type Settings struct {
	URL    string      `json:"url,omitempty"`
	Tenant string      `json:"tenant,omitempty"`
	Vault  string      `json:"vault,omitempty"`
	Auth   *Credential `json:"auth,omitempty"`
}

// Credential is the OAuth login stored by mu login. Its String never shows a token.
type Credential struct {
	Issuer       string    `json:"issuer"`
	ClientID     string    `json:"client_id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scope        string    `json:"scope,omitempty"`
}

// String describes the credential without its tokens.
func (c Credential) String() string {
	return fmt.Sprintf("credential{issuer=%s client=%s expires=%s tokens=[redacted]}",
		c.Issuer, c.ClientID, c.ExpiresAt.Format(time.RFC3339))
}

// Format makes every verb of the fmt family use String, so %+v cannot print a token.
func (c Credential) Format(state fmt.State, _ rune) {
	_, _ = fmt.Fprint(state, c.String())
}

// Set validates value and stores it under key. A URL is normalized first.
func (s *Settings) Set(key, value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return failure.New(failure.Usage, key, "empty value, expected a non empty string")
	}
	if key == KeyURL {
		normalized, err := NormalizeURL(value)
		if err != nil {
			return err
		}
		value = normalized
	}
	field, err := s.field(key)
	if err != nil {
		return err
	}
	*field = value
	return nil
}

// Unset clears key.
func (s *Settings) Unset(key string) error {
	field, err := s.field(key)
	if err != nil {
		return err
	}
	*field = ""
	return nil
}

// Get returns the value stored under key, empty when unset.
func (s Settings) Get(key string) (string, error) {
	field, err := s.field(key)
	if err != nil {
		return "", err
	}
	return *field, nil
}

func (s *Settings) field(key string) (*string, error) {
	switch key {
	case KeyURL:
		return &s.URL, nil
	case KeyTenant:
		return &s.Tenant, nil
	case KeyVault:
		return &s.Vault, nil
	}
	message := fmt.Sprintf("unknown config key %q, expected one of %s", key, strings.Join(Keys, ", "))
	return nil, failure.New(failure.Usage, "", message)
}

// NormalizeURL checks that raw is an absolute http or https URL without query or
// fragment, and drops the trailing slash.
func NormalizeURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	parsed, err := url.Parse(trimmed)
	if err != nil || !validServerURL(parsed) {
		message := fmt.Sprintf("invalid server URL %q, expected http(s)://host[:port][/path]", raw)
		return "", failure.New(failure.Usage, "", message)
	}
	return strings.TrimRight(trimmed, "/"), nil
}

func validServerURL(parsed *url.URL) bool {
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}
	return parsed.Host != "" && parsed.RawQuery == "" && parsed.Fragment == "" && parsed.User == nil
}
