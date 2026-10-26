package cli

import (
	"context"
	"io"

	"github.com/markupp-labs/markupp-cli/internal/client"
)

// GraphReader reads the organization the server derives for a vault: the graph of
// links, the communities and every link. *client.Client satisfies it.
type GraphReader interface {
	VaultGraph(ctx context.Context, vault string, filter client.GraphFilter) (io.ReadCloser, error)
	PublicVaultGraph(ctx context.Context, slug string, filter client.GraphFilter) (io.ReadCloser, error)
	NoteGraph(ctx context.Context, id string, depth int, filter client.GraphFilter) (io.ReadCloser, error)
	VaultCommunities(ctx context.Context, vault string, page client.Page) (io.ReadCloser, error)
	VaultCommunity(ctx context.Context, vault string, cid int64) (io.ReadCloser, error)
	VaultLinks(ctx context.Context, vault string, filter client.LinkFilter, page client.Page) (io.ReadCloser, error)
	PublicVaultLinks(ctx context.Context, slug string, filter client.LinkFilter, page client.Page) (io.ReadCloser, error)
}
