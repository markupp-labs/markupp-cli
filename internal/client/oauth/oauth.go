// Package oauth is the OAuth 2.1 public client of mu: discovery of the authorization
// server (RFC 8414), PKCE (RFC 7636), the authorization code and refresh grants, token
// revocation (RFC 7009) and the loopback redirect of native apps (RFC 8252). It knows
// nothing of the config file; the session package decides when to call it.
package oauth

import (
	"net/http"
	"time"

	"github.com/markupp-labs/markupp-cli/internal/client"
)

// The fixed locations and limits of the flow.
const (
	MetadataPath    = "/.well-known/oauth-authorization-server"
	CallbackPath    = "/callback"
	CallbackTimeout = 5 * time.Minute
	requestTimeout  = 30 * time.Second
	maxBody         = 64 << 10
)

// Config identifies the client to the authorization server.
type Config struct {
	ClientID       string
	Scope          string
	AllowPlainHTTP bool
	// HTTP nil means client.NewHTTPClient(nil, nil): the redirect policy of the API.
	HTTP *http.Client
}

// Client talks to the OAuth endpoints of one or more Markupp servers.
type Client struct {
	cfg  Config
	http *http.Client
}

// New returns a Client for cfg.
func New(cfg Config) *Client {
	httpClient := cfg.HTTP
	if httpClient == nil {
		httpClient = client.NewHTTPClient(nil, nil)
	}
	return &Client{cfg: cfg, http: httpClient}
}

// Metadata is the part of the authorization server metadata mu uses.
type Metadata struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	RevocationEndpoint    string `json:"revocation_endpoint"`
}

// Tokens is what the token endpoint issued. ExpiresIn is relative: the caller turns it
// into a deadline with its own clock.
type Tokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    time.Duration
	Scope        string
}
