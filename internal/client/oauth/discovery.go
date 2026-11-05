package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/failure"
)

const metadataShape = "expected a JSON object with issuer, authorization_endpoint, token_endpoint and revocation_endpoint"

// Discover reads the authorization server metadata at the origin of server. The
// issuer must be that origin (RFC 8414, section 3.3) and every endpoint must live on
// it, so a refresh token never leaves the server it was issued by.
func (c *Client) Discover(ctx context.Context, server string) (Metadata, error) {
	base, err := client.CheckBaseURL(server, c.cfg.AllowPlainHTTP)
	if err != nil {
		return Metadata{}, err
	}
	origin := originOf(base)
	resp, err := c.fetchMetadata(ctx, origin+MetadataPath)
	if err != nil {
		return Metadata{}, err
	}
	defer discard(resp)
	meta, err := decodeMetadata(origin+MetadataPath, resp)
	if err != nil {
		return Metadata{}, err
	}
	return meta, checkOrigins(meta, origin)
}

func (c *Client) fetchMetadata(ctx context.Context, target string) (*http.Response, error) {
	resp, err := c.send(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		defer discard(resp)
		return nil, metadataStatusError(target, resp.StatusCode)
	}
	return resp, nil
}

func metadataStatusError(target string, status int) error {
	message := fmt.Sprintf("server answered %d, expected 200 with the OAuth metadata", status)
	err := failure.New(kindForStatus(status), "GET "+target, message)
	if status == http.StatusNotFound {
		return err.WithHint("the server may not support OAuth login; upgrade it, or set MARKUPP_TOKEN")
	}
	return err
}

func decodeMetadata(target string, resp *http.Response) (Metadata, error) {
	var meta Metadata
	err := json.NewDecoder(limitBody(resp)).Decode(&meta)
	if err != nil || meta.Issuer == "" || meta.TokenEndpoint == "" {
		return Metadata{}, failure.New(failure.Internal, "GET "+target, "malformed OAuth metadata, "+metadataShape).WithCause(err)
	}
	return meta, nil
}

// checkOrigins refuses an issuer other than origin and an endpoint on another origin.
func checkOrigins(meta Metadata, origin string) error {
	if strings.TrimSuffix(meta.Issuer, "/") != origin {
		message := fmt.Sprintf("the OAuth issuer is %q, expected %q, the origin of the server", meta.Issuer, origin)
		return failure.New(failure.Internal, origin+MetadataPath, message)
	}
	endpoints := []struct{ name, value string }{
		{"authorization_endpoint", meta.AuthorizationEndpoint},
		{"token_endpoint", meta.TokenEndpoint},
		{"revocation_endpoint", meta.RevocationEndpoint},
	}
	for _, endpoint := range endpoints {
		if err := checkSameOrigin(endpoint.name, endpoint.value, origin); err != nil {
			return err
		}
	}
	return nil
}

func checkSameOrigin(name, raw, origin string) error {
	if SameOrigin(raw, origin) {
		return nil
	}
	message := fmt.Sprintf("%s is %q, expected a URL on %s", name, raw, origin)
	return failure.New(failure.Internal, origin+MetadataPath, message)
}

// originOf returns scheme://host[:port] of u, lowercased.
func originOf(u *url.URL) string {
	return strings.ToLower(u.Scheme + "://" + u.Host)
}

// SameOrigin reports whether a and b are absolute URLs on the same origin.
func SameOrigin(a, b string) bool {
	first, errA := url.Parse(a)
	second, errB := url.Parse(b)
	if errA != nil || errB != nil || first.Host == "" {
		return false
	}
	return originOf(first) == originOf(second)
}
