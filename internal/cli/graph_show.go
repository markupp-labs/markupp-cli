package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/output"
)

const graphShowLong = `Show the graph of a vault, or with --note the local graph of a note: the
nodes, one per line, then a blank line and the links between them, by node.

A node is a note, with its degree, PageRank and community, or a title that no
note answers to, marked (unresolved). The filters hide orphan notes, unresolved
nodes and inferred links, or keep one community; a link that loses an end goes
too. With --note, the graph holds the notes up to --depth hops away (1 to 3,
default 1) following links both ways.

The vault graph is not paginated: the whole vault comes in one answer. For a
large vault, write it to a file with --json > graph.json and work on the file.

` + publicNeedsServer

// graphShowFlags are the flags of mu graph show.
type graphShowFlags struct {
	vault          vaultFlags
	note           string
	depth          int
	hideOrphans    bool
	hideUnresolved bool
	hideInferred   bool
	community      int64
}

const (
	minDepth = 1
	maxDepth = 3
)

var (
	graphNodeColumns = []output.Column{
		{Name: "ID", Width: idWidth}, {Name: "DEGREE", Width: 6}, {Name: "PAGERANK", Width: 8},
		{Name: "COMMUNITY", Width: 9}, {Name: "PATH"},
	}
	graphLinkColumns = []output.Column{{Name: "SOURCE", Width: pathWidth}, {Name: "TYPE", Width: typeWidth}, {Name: "TARGET"}}
)

func newGraphShowCmd(app *App) *cobra.Command {
	var flags graphShowFlags
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show the graph of a vault, or of a note with --note",
		Long:  graphShowLong,
		Example: `  mu graph show --note 3f2a9c1e-5b7d-4e8a-9c0f-1a2b3c4d5e6f --depth 2 --no-unresolved
  mu graph show --vault 3b241101-e2bb-4255-8caf-4136c566a962 --no-orphans --json > graph.json`,
		Args: exactArgs(),
		RunE: func(cmd *cobra.Command, _ []string) error { return app.showGraph(cmd, flags) },
	}
	addGraphShowFlags(cmd, &flags)
	return cmd
}

func addGraphShowFlags(cmd *cobra.Command, flags *graphShowFlags) {
	addVaultFlags(cmd, &flags.vault, true)
	set := cmd.Flags()
	set.StringVar(&flags.note, "note", "", "show the local graph of this note id instead of the vault")
	set.IntVar(&flags.depth, "depth", 0, fmt.Sprintf("hops from --note, %d to %d (server default %d)", minDepth, maxDepth, minDepth))
	set.BoolVar(&flags.hideOrphans, "no-orphans", false, "hide the notes without links")
	set.BoolVar(&flags.hideUnresolved, "no-unresolved", false, "hide the nodes of titles no note answers to")
	set.BoolVar(&flags.hideInferred, "no-inferred", false, "hide the links the indexer inferred")
	set.Int64Var(&flags.community, "community", 0, "keep only the notes of this community id")
}

func (a *App) showGraph(cmd *cobra.Command, flags graphShowFlags) error {
	if err := checkGraphScope(cmd, flags); err != nil {
		return err
	}
	r, err := a.renderer(false)
	if err != nil {
		return err
	}
	filter := graphFilter(flags, cmd.Flags().Changed("community"))
	body, err := a.fetchGraph(cmd.Context(), flags, filter)
	if err != nil {
		return err
	}
	return r.object(body, renderGraph)
}

// checkGraphScope refuses --note with a vault flag, and a --depth without --note or
// out of range.
func checkGraphScope(cmd *cobra.Command, flags graphShowFlags) error {
	depthSet := cmd.Flags().Changed("depth")
	if flags.note == "" && depthSet {
		return failure.New(failure.Usage, "--depth", fmt.Sprintf("--depth %d needs --note, expected a note id", flags.depth))
	}
	if flags.note != "" && (flags.vault.vault != "" || flags.vault.public != "") {
		message := fmt.Sprintf("--note %q shows the graph of a note, expected no --vault or --public", flags.note)
		return failure.New(failure.Usage, "--note", message)
	}
	if depthSet && (flags.depth < minDepth || flags.depth > maxDepth) {
		message := fmt.Sprintf("invalid value %d, expected %d to %d", flags.depth, minDepth, maxDepth)
		return failure.New(failure.Usage, "--depth", message)
	}
	return nil
}

func graphFilter(flags graphShowFlags, communitySet bool) client.GraphFilter {
	filter := client.GraphFilter{
		HideOrphans: flags.hideOrphans, HideUnresolved: flags.hideUnresolved, HideInferred: flags.hideInferred,
	}
	if communitySet {
		community := flags.community
		filter.Community = &community
	}
	return filter
}

func (a *App) fetchGraph(ctx context.Context, flags graphShowFlags, filter client.GraphFilter) (io.ReadCloser, error) {
	if flags.note != "" {
		return a.fetchNoteGraph(ctx, flags, filter)
	}
	target, err := a.resolveVault(flags.vault)
	if err != nil {
		return nil, err
	}
	graph, err := a.graphConnect(ctx, target)
	if err != nil {
		return nil, err
	}
	if target.PublicSlug != "" {
		return graph.PublicVaultGraph(ctx, target.PublicSlug, filter)
	}
	return graph.VaultGraph(ctx, target.ID, filter)
}

func (a *App) fetchNoteGraph(ctx context.Context, flags graphShowFlags, filter client.GraphFilter) (io.ReadCloser, error) {
	svc, err := a.connect(ctx, Target{})
	if err != nil {
		return nil, err
	}
	return svc.Graph.NoteGraph(ctx, flags.note, flags.depth, filter)
}
