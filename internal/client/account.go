package client

import (
	"context"
	"io"
	"net/http"
)

// GetHealth reads GET /healthz, the probe that needs no token. A 503 is an error of
// kind Unavailable.
func (c *Client) GetHealth(ctx context.Context) (io.ReadCloser, error) {
	return c.open(ctx, request{method: http.MethodGet, path: endpoint("healthz")})
}

// GetAuthenticatedUser reads GET /auth/me, the account of the token.
func (c *Client) GetAuthenticatedUser(ctx context.Context) (io.ReadCloser, error) {
	return c.open(ctx, request{method: http.MethodGet, path: endpoint("auth", "me")})
}
