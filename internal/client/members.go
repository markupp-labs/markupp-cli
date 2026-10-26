package client

import (
	"context"
	"io"
)

func (c *Client) ListMembers(context.Context) (io.ReadCloser, error) { return nil, errAdminStub }

func (c *Client) ChangeMemberRole(context.Context, string, string) error { return errAdminStub }

func (c *Client) RemoveMember(context.Context, string) error { return errAdminStub }

func (c *Client) InviteMember(context.Context, string, string) (io.ReadCloser, error) {
	return nil, errAdminStub
}

func (c *Client) ListTenantInvites(context.Context) (io.ReadCloser, error) {
	return nil, errAdminStub
}

func (c *Client) RevokeInvite(context.Context, string) error { return errAdminStub }
