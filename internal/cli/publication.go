package cli

import (
	"context"
	"io"
)

// VaultPublisher reads, sets and removes the publication of a vault.
type VaultPublisher interface {
	GetPublication(ctx context.Context, id string) (io.ReadCloser, error)
	PublishVault(ctx context.Context, id, slug, title string) (io.ReadCloser, error)
	UnpublishVault(ctx context.Context, id string) error
}
