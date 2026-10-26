package cli

import (
	"context"
	"encoding/json"
	"io"

	"github.com/spf13/cobra"

	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/output"
)

// NoteLinkReader reads what connects one note to the others: its links, its unlinked
// mentions and the links worth adding. *client.Client satisfies it.
type NoteLinkReader interface {
	NoteLinks(ctx context.Context, id, linkType string, page client.Page) (io.ReadCloser, error)
	NoteBacklinks(ctx context.Context, id, linkType string, page client.Page) (io.ReadCloser, error)
	NoteMentions(ctx context.Context, id string, page client.Page) (io.ReadCloser, error)
	NoteBackmentions(ctx context.Context, id string, page client.Page) (io.ReadCloser, error)
	NoteSuggestions(ctx context.Context, id string, page client.Page) (io.ReadCloser, error)
}

const notesLinksLong = `List the links a note makes, one per line: the note at the other end, the
link type and the title as cited. With --back, list the links other notes of the
vault make to this one.

A link is a wikilink, an embed, a markdown link to a .md file, or a typed link
(a frontmatter key whose value is a wikilink, shown as typed:KEY). An unresolved
link has no note at the other end and shows "-"; the indexer's guess for it is a
separate link of type inferred.`

// noteLinksFlags are the flags of mu notes links.
type noteLinksFlags struct {
	list     listFlags
	back     bool
	linkType string
}

var noteLinkColumns = []output.Column{
	{Name: "ID", Width: idWidth}, {Name: "TYPE", Width: typeWidth}, {Name: "PATH", Width: pathWidth}, {Name: "CITED"},
}

func newNotesLinksCmd(app *App) *cobra.Command {
	var flags noteLinksFlags
	cmd := &cobra.Command{
		Use:   "links ID",
		Short: "List the links of a note, or with --back the links to it",
		Long:  notesLinksLong,
		Example: `  mu notes links 3f2a9c1e-5b7d-4e8a-9c0f-1a2b3c4d5e6f --type typed
  mu notes links 3f2a9c1e-5b7d-4e8a-9c0f-1a2b3c4d5e6f --back --json | jq -r '.[].source.id'`,
		Args: exactArgs("ID"),
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.listNoteLinks(cmd.Context(), args[0], flags)
		},
	}
	addListFlags(cmd, &flags.list)
	cmd.Flags().BoolVar(&flags.back, "back", false, "list the links from other notes to this one (backlinks)")
	cmd.Flags().StringVar(&flags.linkType, "type", "", linkTypeHelp)
	return cmd
}

func (a *App) listNoteLinks(ctx context.Context, id string, flags noteLinksFlags) error {
	if err := checkChoice("--type", flags.linkType, linkTypes); err != nil {
		return err
	}
	page, r, err := a.startListing(flags.list)
	if err != nil {
		return err
	}
	svc, err := a.connect(ctx, Target{})
	if err != nil {
		return err
	}
	body, err := fetchNoteLinks(ctx, svc.NoteLinks, id, flags, page)
	if err != nil {
		return err
	}
	return r.list(body, noteLinkColumns, noteLinkRow(flags.back))
}

func fetchNoteLinks(ctx context.Context, links NoteLinkReader, id string, flags noteLinksFlags, page client.Page) (io.ReadCloser, error) {
	if flags.back {
		return links.NoteBacklinks(ctx, id, flags.linkType, page)
	}
	return links.NoteLinks(ctx, id, flags.linkType, page)
}

// noteLinkRow shows the note at the other end of each link: the target, or the
// source for backlinks.
func noteLinkRow(back bool) rowFunc {
	return func(raw json.RawMessage) ([]string, error) {
		link, err := decodeLink(raw)
		if err != nil {
			return nil, err
		}
		other := summaryOrEmpty(link.Target)
		if back {
			other = link.Source
		}
		return []string{other.ID, linkTypeLabel(link.Type, link.Predicate), other.Path, link.CitedTitle}, nil
	}
}
