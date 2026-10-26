package client

import (
	"context"
	"io"
	"net/http"
)

// Notifications returns the notification feed of the account: the items that ask for
// its attention and how many of them are unread.
func (c *Client) Notifications(ctx context.Context) (io.ReadCloser, error) {
	return c.open(ctx, request{method: http.MethodGet, path: endpoint("me", "notifications")})
}
