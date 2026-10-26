package client

import (
	"context"
	"io"
	"net/http"

	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// ListMembers returns the MemberResponse array of the members of the tenant.
func (c *Client) ListMembers(ctx context.Context) (io.ReadCloser, error) {
	return c.open(ctx, request{method: http.MethodGet, path: endpoint("users")})
}

// ChangeMemberRole gives the member with the account id the role.
func (c *Client) ChangeMemberRole(ctx context.Context, id, role string) error {
	body, err := encodeAdminBody(schema.RoleRequest{Role: role})
	if err != nil {
		return err
	}
	return c.exec(ctx, request{method: http.MethodPatch, path: endpoint("users", id), body: body})
}

// RemoveMember removes the member with the account id from the tenant.
func (c *Client) RemoveMember(ctx context.Context, id string) error {
	return c.exec(ctx, request{method: http.MethodDelete, path: endpoint("users", id)})
}

// InviteMember invites email to the tenant with role and returns the
// IssuedInviteResponse.
func (c *Client) InviteMember(ctx context.Context, email, role string) (io.ReadCloser, error) {
	body, err := encodeAdminBody(schema.InviteRequest{Email: email, Role: role})
	if err != nil {
		return nil, err
	}
	return c.open(ctx, request{method: http.MethodPost, path: endpoint("users", "invites"), body: body})
}

// ListTenantInvites returns the InviteResponse array of the invites of the tenant.
func (c *Client) ListTenantInvites(ctx context.Context) (io.ReadCloser, error) {
	return c.open(ctx, request{method: http.MethodGet, path: endpoint("users", "invites")})
}

// RevokeInvite revokes the unanswered invite id of the tenant.
func (c *Client) RevokeInvite(ctx context.Context, id string) error {
	return c.exec(ctx, request{method: http.MethodDelete, path: endpoint("users", "invites", id)})
}
