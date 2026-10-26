// Package session keeps the OAuth login of mu: it logs in through the browser, stores
// the credential in the config file, hands the access token to the API client and
// refreshes it under the config lock, so concurrent mu processes rotate the single use
// refresh token once.
package session

import (
	"context"
	"net/http"
	"time"

	"github.com/markupp-labs/markupp-cli/internal/client/oauth"
	"github.com/markupp-labs/markupp-cli/internal/config"
)

// The OAuth identity of mu: the built in public client of the server and the scopes
// every command may need.
const (
	ClientID = "markupp-cli"
	Scope    = "notas:leitura notas:escrita administracao"
)

// ExpiryMargin is how long before its expiry a stored access token is refreshed, so a
// token never expires between the check and the request.
const ExpiryMargin = 30 * time.Second

// Clock returns the current time.
type Clock func() time.Time

// Store is the config file as the session sees it. *config.Store satisfies it: Save
// writes a synced temporary file and renames it, and Lock is the cross process lock.
type Store interface {
	Path() (string, error)
	Load() (config.Settings, error)
	Save(config.Settings) error
	Lock(ctx context.Context) (unlock func() error, err error)
}

// NewOAuth returns the OAuth client of mu. httpClient nil means the API transport.
func NewOAuth(httpClient *http.Client, allowPlainHTTP bool) *oauth.Client {
	return oauth.New(oauth.Config{ClientID: ClientID, Scope: Scope, AllowPlainHTTP: allowPlainHTTP, HTTP: httpClient})
}
