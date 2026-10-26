package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/output"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

const graphCommunitiesLong = `List the communities of the notes of a vault, largest first, each with its
hub: the note of highest PageRank. With CID, show that one community.

Communities come from the links between notes; a note without links has none.
The indexer recomputes them when links or titles change, and a community keeps
its id while it keeps most of its notes.`

// graphCommunitiesFlags are the flags of mu graph communities.
type graphCommunitiesFlags struct {
	vault vaultFlags
	list  listFlags
}

var communityColumns = []output.Column{
	{Name: "ID", Width: 6}, {Name: "SIZE", Width: 6}, {Name: "HUB", Width: idWidth}, {Name: "PATH"},
}

func newGraphCommunitiesCmd(app *App) *cobra.Command {
	var flags graphCommunitiesFlags
	cmd := &cobra.Command{
		Use:   "communities [CID]",
		Short: "List the communities of a vault, or show one",
		Long:  graphCommunitiesLong,
		Example: `  mu graph communities --vault 3b241101-e2bb-4255-8caf-4136c566a962 --limit 5
  mu graph communities 7 --json | jq -r .hub.id`,
		Args: optionalArg("CID"),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return app.showCommunity(cmd, args[0], flags)
			}
			return app.listCommunities(cmd.Context(), flags)
		},
	}
	addVaultFlags(cmd, &flags.vault, false)
	addListFlags(cmd, &flags.list)
	return cmd
}

// optionalArg accepts no positional argument or the one named name.
func optionalArg(name string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) <= 1 {
			return nil
		}
		message := fmt.Sprintf("expected at most 1 argument (%s), got %d: %q", name, len(args), args)
		return failure.New(failure.Usage, cmd.CommandPath(), message)
	}
}

func (a *App) listCommunities(ctx context.Context, flags graphCommunitiesFlags) error {
	page, r, err := a.startListing(flags.list)
	if err != nil {
		return err
	}
	vault, graph, err := a.vaultGraphReader(ctx, flags.vault)
	if err != nil {
		return err
	}
	body, err := graph.VaultCommunities(ctx, vault, page)
	if err != nil {
		return err
	}
	return r.list(body, communityColumns, communityRow)
}

func (a *App) showCommunity(cmd *cobra.Command, arg string, flags graphCommunitiesFlags) error {
	cid, err := parseCommunityID(cmd, arg)
	if err != nil {
		return err
	}
	r, err := a.renderer(flags.list.ndjson)
	if err != nil {
		return err
	}
	vault, graph, err := a.vaultGraphReader(cmd.Context(), flags.vault)
	if err != nil {
		return err
	}
	body, err := graph.VaultCommunity(cmd.Context(), vault, cid)
	if err != nil {
		return err
	}
	return r.object(body, renderCommunity)
}

// parseCommunityID reads CID, refusing the paging flags that only apply to the listing.
func parseCommunityID(cmd *cobra.Command, arg string) (int64, error) {
	if cmd.Flags().Changed("limit") || cmd.Flags().Changed("offset") {
		message := fmt.Sprintf("--limit and --offset page the listing, expected neither with CID %q", arg)
		return 0, failure.New(failure.Usage, cmd.CommandPath(), message)
	}
	cid, err := strconv.ParseInt(arg, 10, 64)
	if err != nil {
		message := fmt.Sprintf("invalid community id %q, expected an integer", arg)
		return 0, failure.New(failure.Usage, "CID", message)
	}
	return cid, nil
}

// vaultGraphReader resolves the vault of the account and connects to it.
func (a *App) vaultGraphReader(ctx context.Context, flags vaultFlags) (string, GraphReader, error) {
	target, err := a.resolveVault(flags)
	if err != nil {
		return "", nil, err
	}
	graph, err := a.graphConnect(ctx, target)
	return target.ID, graph, err
}

func decodeCommunity(raw json.RawMessage) (schema.Community, error) {
	var community schema.Community
	if err := json.Unmarshal(raw, &community); err != nil {
		return schema.Community{}, malformedRecord(raw, err)
	}
	return community, nil
}

func communityRow(raw json.RawMessage) ([]string, error) {
	community, err := decodeCommunity(raw)
	if err != nil {
		return nil, err
	}
	id, size := strconv.FormatInt(community.ID, 10), strconv.Itoa(community.Size)
	return []string{id, size, community.Hub.ID, community.Hub.Path}, nil
}

func renderCommunity(p *output.Printer, raw json.RawMessage) error {
	community, err := decodeCommunity(raw)
	if err != nil {
		return err
	}
	return p.Fields(
		output.Field{Key: "id", Value: strconv.FormatInt(community.ID, 10)},
		output.Field{Key: "size", Value: strconv.Itoa(community.Size)},
		output.Field{Key: "hub", Value: community.Hub.ID},
		output.Field{Key: "hub path", Value: community.Hub.Path},
		output.Field{Key: "hub title", Value: community.Hub.Title},
	)
}
