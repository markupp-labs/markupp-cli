package client

import (
	"context"
	"io"
	"net/http"
)

// ListMyInvites returns the PendingInviteResponse array of the invites addressed to
// the account.
func (c *Client) ListMyInvites(ctx context.Context) (io.ReadCloser, error) {
	return c.open(ctx, request{method: http.MethodGet, path: endpoint("me", "invites")})
}

// AcceptInvite joins the tenant of the invite id and returns its TenantResponse.
func (c *Client) AcceptInvite(ctx context.Context, id string) (io.ReadCloser, error) {
	return c.open(ctx, request{method: http.MethodPost, path: endpoint("me", "invites", id, "accept")})
}

// DeclineInvite declines the invite id.
func (c *Client) DeclineInvite(ctx context.Context, id string) error {
	return c.exec(ctx, request{method: http.MethodPost, path: endpoint("me", "invites", id, "decline")})
}
