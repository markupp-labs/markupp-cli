package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/config"
	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/output"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

var searchColumns = []output.Column{{Name: "ID", Width: 36}, {Name: "PATH", Width: 40}, {Name: "SNIPPET"}}

const notesSearchLong = `Search notes by words in the title, aliases and content, most relevant first.

QUERY uses web search syntax: loose words must all match, "quotes" ask for the
phrase, OR accepts either side and -word excludes. It is not a substring search.
The search covers the vault given by --vault, ` + config.EnvVault + ` or the config file,
and every vault of the account when none is set. Notes appear once indexed.

The results are paginated: --limit is clamped to 100 and --offset to 10000 by the
server. They carry a snippet with the matches in [[ ]], never the note content:
fetching content is a second call per result, with mu notes get.

--public SLUG searches a published vault instead, without credentials.
` + publicNeedsServer

const notesSearchExample = `  mu notes search kubernetes --limit 5
  mu notes search '"exact phrase" OR term -draft' --vault 3b241101-e2bb-4255-8caf-4136c566a962
  mu notes search kubernetes --limit 3 --ndjson | jq -r .note.id | xargs -n1 mu notes get
  mu notes search kubernetes --public my-garden --limit 5`

type notesSearchCmd struct {
	app   *App
	vault vaultFlags
	list  listFlags
}

func newNotesSearchCmd(app *App) *cobra.Command {
	c := &notesSearchCmd{app: app}
	cmd := &cobra.Command{
		Use:     "search QUERY",
		Short:   "Search notes by text, without content",
		Long:    notesSearchLong,
		Example: notesSearchExample,
		Args:    exactArgs("QUERY"),
		RunE:    func(cmd *cobra.Command, args []string) error { return c.run(cmd, args[0]) },
	}
	addVaultFlags(cmd, &c.vault, true)
	addListFlags(cmd, &c.list)
	return cmd
}

func (c *notesSearchCmd) run(cmd *cobra.Command, text string) error {
	page, r, err := c.request(text)
	if err != nil {
		return err
	}
	body, err := c.fetch(cmd.Context(), text, page)
	if err != nil {
		return err
	}
	return r.list(body, searchColumns, searchRow)
}

// request validates the query and the flags before anything reaches the server.
func (c *notesSearchCmd) request(text string) (client.Page, renderer, error) {
	if strings.TrimSpace(text) == "" {
		return client.Page{}, renderer{}, failure.New(failure.Usage, "QUERY", fmt.Sprintf("empty query %q, expected words to search", text))
	}
	if err := c.vault.exclusive(); err != nil {
		return client.Page{}, renderer{}, err
	}
	return c.app.startListing(c.list)
}

// fetch searches the public vault of --public, or the optional vault of the account.
func (c *notesSearchCmd) fetch(ctx context.Context, text string, page client.Page) (io.ReadCloser, error) {
	if c.vault.public != "" {
		public, err := c.app.publicReader(ctx)
		if err != nil {
			return nil, err
		}
		return publicRead(public.PublicSearch(ctx, c.vault.public, text, page))
	}
	resolved, _, err := c.app.resolve(c.vault.vault)
	if err != nil {
		return nil, err
	}
	notes, err := c.app.notes(ctx, c.vault.vault)
	if err != nil {
		return nil, err
	}
	return notes.SearchNotes(ctx, text, resolved.Vault.Value, page)
}

func searchRow(raw json.RawMessage) ([]string, error) {
	hit, err := decodeRecord[schema.SearchHit](raw)
	if err != nil {
		return nil, err
	}
	return []string{hit.Note.ID, hit.Note.Path, output.Snippet(hit.Snippet)}, nil
}
