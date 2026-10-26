package client

import (
	"context"
	"errors"
	"io"
)

// GetHealth reads the health probe.
func (c *Client) GetHealth(context.Context) (io.ReadCloser, error) {
	return nil, errors.New("not implemented")
}

// GetAuthenticatedUser reads the account.
func (c *Client) GetAuthenticatedUser(context.Context) (io.ReadCloser, error) {
	return nil, errors.New("not implemented")
}

// ListMyTenants lists the tenants of the account.
func (c *Client) ListMyTenants(context.Context) (io.ReadCloser, error) {
	return nil, errors.New("not implemented")
}
