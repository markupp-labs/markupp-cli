package client

import (
	"context"
	"io"
)

func (c *Client) ListMyInvites(context.Context) (io.ReadCloser, error) { return nil, errAdminStub }

func (c *Client) AcceptInvite(context.Context, string) (io.ReadCloser, error) {
	return nil, errAdminStub
}

func (c *Client) DeclineInvite(context.Context, string) error { return errAdminStub }
