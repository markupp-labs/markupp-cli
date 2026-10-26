package oauth

import "context"

// ExchangeCode trades the authorization code for tokens.
func (c *Client) ExchangeCode(context.Context, Metadata, string, string, string) (Tokens, error) {
	return Tokens{}, errNotImplemented
}

// Refresh rotates the refresh token.
func (c *Client) Refresh(context.Context, Metadata, string) (Tokens, error) {
	return Tokens{}, errNotImplemented
}

// IsInvalidGrant reports a refused grant.
func IsInvalidGrant(error) bool { return false }
