// Package config locates, reads and writes the mu configuration file and resolves
// each setting from a flag, the environment or the file, in that order.
package config

import (
	"context"
	"time"
)

// LookupEnv reads an environment variable, like os.LookupEnv.
type LookupEnv func(key string) (string, bool)

// Environment variables read by mu.
const (
	EnvURL    = "MARKUPP_URL"
	EnvToken  = "MARKUPP_TOKEN"
	EnvTenant = "MARKUPP_TENANT"
	EnvVault  = "MARKUPP_VAULT"
)

// Settable keys of the configuration file.
const (
	KeyURL    = "url"
	KeyTenant = "tenant"
	KeyVault  = "vault"
)

// Keys lists the keys mu config set and unset accept, in display order.
var Keys = []string{KeyURL, KeyTenant, KeyVault}

// StaleAfter is the age from which a lock file is taken as left behind by a dead process.
const StaleAfter = 30 * time.Second

// Settings is the content of the configuration file.
type Settings struct {
	URL    string      `json:"url,omitempty"`
	Tenant string      `json:"tenant,omitempty"`
	Vault  string      `json:"vault,omitempty"`
	Auth   *Credential `json:"auth,omitempty"`
}

// Credential is the OAuth login stored by mu login.
type Credential struct {
	Issuer       string    `json:"issuer"`
	ClientID     string    `json:"client_id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scope        string    `json:"scope,omitempty"`
}

// Source says where a resolved value came from.
type Source string

// The sources, from the strongest to the weakest. FromNone means unset.
const (
	FromNone Source = ""
	FromFlag Source = "flag"
	FromEnv  Source = "env"
	FromFile Source = "file"
)

// Value is a resolved setting and its source.
type Value struct {
	Value  string
	Source Source
}

// Overrides are the values given by flags.
type Overrides struct {
	URL    string
	Tenant string
	Vault  string
}

// Resolved holds every setting after precedence. Token only ever comes from the
// environment; the stored login is Settings.Auth.
type Resolved struct {
	URL    Value
	Tenant Value
	Vault  Value
	Token  Value
}

// Store reads and writes the configuration file.
type Store struct {
	path    string
	pathErr error
	now     func() time.Time
}

func Path(lookup LookupEnv) (string, error)                             { return "", nil }
func NormalizeURL(raw string) (string, error)                           { return raw, nil }
func (s *Settings) Set(key, value string) error                         { return nil }
func (s *Settings) Unset(key string) error                              { return nil }
func (s Settings) Get(key string) (string, error)                       { return "", nil }
func (c Credential) String() string                                     { return "" }
func Resolve(flags Overrides, lookup LookupEnv, file Settings) Resolved { return Resolved{} }
func NewStore(path string) *Store                                       { return &Store{path: path, now: time.Now} }
func DefaultStore(lookup LookupEnv) *Store                              { return &Store{now: time.Now} }
func (s *Store) Path() (string, error)                                  { return s.path, s.pathErr }
func (s *Store) Load() (Settings, error)                                { return Settings{}, nil }
func (s *Store) Save(settings Settings) error                           { return nil }
func (s *Store) Lock(ctx context.Context) (func() error, error) {
	return func() error { return nil }, nil
}
func (s *Store) Update(ctx context.Context, mutate func(*Settings) error) error { return nil }
