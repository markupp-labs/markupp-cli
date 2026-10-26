package cli

import (
	"context"
	"io"
)

// VaultAdmin lists, creates, renames and deletes the vaults of the tenant.
type VaultAdmin interface {
	ListVaults(ctx context.Context) (io.ReadCloser, error)
	CreateVault(ctx context.Context, name string) (io.ReadCloser, error)
	RenameVault(ctx context.Context, id, name string) (io.ReadCloser, error)
	DeleteVault(ctx context.Context, id string) error
}

// VaultMembership lists, adds and removes the members of a vault.
type VaultMembership interface {
	ListVaultMembers(ctx context.Context, id string) (io.ReadCloser, error)
	AddVaultMember(ctx context.Context, id, userID string) (io.ReadCloser, error)
	RemoveVaultMember(ctx context.Context, id, userID string) error
}
