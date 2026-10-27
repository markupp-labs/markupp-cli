package cli

import (
	"context"
	"encoding/json"
	"io"

	"github.com/spf13/cobra"

	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/output"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

const graphLinksLong = `List the links of every note of a vault, one per line, by the path of the
note that links: the source, the link type, the target and the title as cited.
An unresolved link shows "-" as target; --status unresolved lists only those,
with the indexer's guess in inferred_target of the --json output.

A vault can hold many links: page with --limit and --offset, or write the
--json output to a file and work on it with jq.

` + publicNeedsServer

// graphLinksFlags are the flags of mu graph links.
type graphLinksFlags struct {
	vault  vaultFlags
	list   listFlags
	filter client.LinkFilter
}

var vaultLinkColumns = []output.Column{
	{Name: "SOURCE", Width: pathWidth}, {Name: "TYPE", Width: typeWidth}, {Name: "TARGET", Width: pathWidth}, {Name: "CITED"},
}

func newGraphLinksCmd(app *App) *cobra.Command {
	var flags graphLinksFlags
	cmd := &cobra.Command{
		Use:   "links",
		Short: "List the links of the notes of a vault",
		Long:  graphLinksLong,
		Example: `  mu graph links --vault 3b241101-e2bb-4255-8caf-4136c566a962 --status unresolved --limit 20
  mu graph links --public my-garden --type typed --json > links.json`,
		Args: exactArgs(),
		RunE: func(cmd *cobra.Command, _ []string) error { return app.listVaultLinks(cmd.Context(), flags) },
	}
	addVaultFlags(cmd, &flags.vault, true)
	addListFlags(cmd, &flags.list)
	cmd.Flags().StringVar(&flags.filter.Status, "status", "", "only resolved or only unresolved links")
	cmd.Flags().StringVar(&flags.filter.Type, "type", "", linkTypeHelp)
	return cmd
}

func (a *App) listVaultLinks(ctx context.Context, flags graphLinksFlags) error {
	if err := checkLinkFilter(flags.filter); err != nil {
		return err
	}
	page, r, err := a.startListing(flags.list)
	if err != nil {
		return err
	}
	body, err := a.fetchVaultLinks(ctx, flags, page)
	if err != nil {
		return err
	}
	return r.list(body, vaultLinkColumns, vaultLinkRow)
}

func checkLinkFilter(filter client.LinkFilter) error {
	if err := checkChoice("--status", filter.Status, linkStatuses); err != nil {
		return err
	}
	return checkChoice("--type", filter.Type, linkTypes)
}

func (a *App) fetchVaultLinks(ctx context.Context, flags graphLinksFlags, page client.Page) (io.ReadCloser, error) {
	graph, target, err := a.graphConnect(ctx, flags.vault)
	if err != nil {
		return nil, err
	}
	if target.PublicSlug != "" {
		return graph.PublicVaultLinks(ctx, target.PublicSlug, flags.filter, page)
	}
	return graph.VaultLinks(ctx, target.ID, flags.filter, page)
}

func vaultLinkRow(raw json.RawMessage) ([]string, error) {
	link, err := decodeRecord[schema.Link](raw)
	if err != nil {
		return nil, err
	}
	target := summaryOrEmpty(link.Target)
	return []string{link.Source.Path, linkTypeLabel(link.Type, link.Predicate), target.Path, link.CitedTitle}, nil
}
