package session

import (
	"context"
	"errors"

	"github.com/markupp-labs/markupp-cli/internal/client/oauth"
)

var errNotImplemented = errors.New("not implemented")

// Source is the client.TokenSource of a stored login.
type Source struct{}

// NewSource returns the Source of the login stored for server.
func NewSource(Store, *oauth.Client, string, Clock) *Source { return &Source{} }

// Token returns the access token.
func (s *Source) Token(context.Context) (string, error) { return "", errNotImplemented }

// Rejected refreshes a rejected token.
func (s *Source) Rejected(context.Context, string) (string, bool, error) {
	return "", false, errNotImplemented
}
