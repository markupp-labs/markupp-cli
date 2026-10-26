package cli

import (
	"context"
	"io"

	"github.com/spf13/cobra"

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

const publicNeedsServer = "--public requires a server with public vaults."

func newGraphCmd(app *App) *cobra.Command {
	cmd := newGroup("graph", "Read the graph, communities and links of a vault",
		"  mu graph show --note 3f2a9c1e-5b7d-4e8a-9c0f-1a2b3c4d5e6f --depth 2\n  mu graph communities --limit 5")
	cmd.AddCommand(newGraphShowCmd(app), newGraphCommunitiesCmd(app), newGraphLinksCmd(app))
	return cmd
}

// graphConnect connects as the account for a vault id, or anonymously for the slug
// of a public vault.
func (a *App) graphConnect(ctx context.Context, target vaultTarget) (GraphReader, error) {
	svc, err := a.connect(ctx, Target{Vault: target.ID, Public: target.PublicSlug != ""})
	if err != nil {
		return nil, err
	}
	return svc.Graph, nil
}
