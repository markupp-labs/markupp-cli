package client

import (
	"context"
	"errors"
	"io"
)

var errVaultsPending = errors.New("not implemented")

// ListVaults returns the vaults of the account in the tenant, a JSON array.
func (c *Client) ListVaults(context.Context) (io.ReadCloser, error) { return nil, errVaultsPending }

// CreateVault creates a vault named name and returns it.
func (c *Client) CreateVault(context.Context, string) (io.ReadCloser, error) {
	return nil, errVaultsPending
}

// RenameVault renames the vault id and returns it.
func (c *Client) RenameVault(context.Context, string, string) (io.ReadCloser, error) {
	return nil, errVaultsPending
}

// DeleteVault deletes the vault id with its notes.
func (c *Client) DeleteVault(context.Context, string) error { return errVaultsPending }

// ListVaultMembers returns the members of the vault id, a JSON array.
func (c *Client) ListVaultMembers(context.Context, string) (io.ReadCloser, error) {
	return nil, errVaultsPending
}

// AddVaultMember adds the tenant account userID to the vault id and returns the member.
func (c *Client) AddVaultMember(context.Context, string, string) (io.ReadCloser, error) {
	return nil, errVaultsPending
}

// RemoveVaultMember removes the account userID from the vault id.
func (c *Client) RemoveVaultMember(context.Context, string, string) error { return errVaultsPending }
