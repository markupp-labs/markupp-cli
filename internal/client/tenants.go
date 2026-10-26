package client

import (
	"context"
	"errors"
	"io"
)

var errAdminStub = errors.New("not implemented")

func (c *Client) ListMyTenants(context.Context) (io.ReadCloser, error) { return nil, errAdminStub }

func (c *Client) CreateTenant(context.Context, string) (io.ReadCloser, error) {
	return nil, errAdminStub
}
