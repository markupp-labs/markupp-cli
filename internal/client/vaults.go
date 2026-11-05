package client

import (
	"context"
	"io"
	"net/http"

	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// ListVaults returns the vaults of the account in the tenant, a JSON array.
func (c *Client) ListVaults(ctx context.Context) (io.ReadCloser, error) {
	return c.open(ctx, request{method: http.MethodGet, path: endpoint("vaults")})
}

// CreateVault creates a vault named name and returns it.
func (c *Client) CreateVault(ctx context.Context, name string) (io.ReadCloser, error) {
	r := request{method: http.MethodPost, path: endpoint("vaults")}
	return openJSON(ctx, c, r, schema.VaultNameRequest{Name: name})
}

// RenameVault renames the vault id and returns it.
func (c *Client) RenameVault(ctx context.Context, id, name string) (io.ReadCloser, error) {
	r := request{method: http.MethodPatch, path: endpoint("vaults", id)}
	return openJSON(ctx, c, r, schema.VaultNameRequest{Name: name})
}

// DeleteVault deletes the vault id with its notes.
func (c *Client) DeleteVault(ctx context.Context, id string) error {
	return c.exec(ctx, request{method: http.MethodDelete, path: endpoint("vaults", id)})
}

// ListVaultMembers returns the members of the vault id, a JSON array.
func (c *Client) ListVaultMembers(ctx context.Context, id string) (io.ReadCloser, error) {
	return c.open(ctx, request{method: http.MethodGet, path: endpoint("vaults", id, "members")})
}

// AddVaultMember adds the tenant account userID to the vault id and returns the member.
func (c *Client) AddVaultMember(ctx context.Context, id, userID string) (io.ReadCloser, error) {
	r := request{method: http.MethodPost, path: endpoint("vaults", id, "members")}
	return openJSON(ctx, c, r, schema.VaultMemberRequest{UserID: userID})
}

// RemoveVaultMember removes the account userID from the vault id. Removing the
// account itself is leaving the vault.
func (c *Client) RemoveVaultMember(ctx context.Context, id, userID string) error {
	return c.exec(ctx, request{method: http.MethodDelete, path: endpoint("vaults", id, "members", userID)})
}
