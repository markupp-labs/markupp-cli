package oauth

import (
	"net/url"
	"strings"
)

// AuthorizeURL returns the URL the browser opens to start the authorization code flow
// with PKCE, asking for the configured scope on the issuer as the resource (RFC 8707).
func (c *Client) AuthorizeURL(meta Metadata, redirectURI string, pkce PKCE, state string) string {
	query := url.Values{
		"response_type":         {"code"},
		"client_id":             {c.cfg.ClientID},
		"redirect_uri":          {redirectURI},
		"code_challenge":        {pkce.Challenge},
		"code_challenge_method": {"S256"},
		"state":                 {state},
		"scope":                 {c.cfg.Scope},
		"resource":              {meta.Issuer},
	}
	separator := "?"
	if strings.Contains(meta.AuthorizationEndpoint, "?") {
		separator = "&"
	}
	return meta.AuthorizationEndpoint + separator + query.Encode()
}
