package cli

import (
	"encoding/json"
	"fmt"
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
fetching content is a second call per result, with mu notes get.`

const notesSearchExample = `  mu notes search kubernetes --limit 5
  mu notes search '"exact phrase" OR term -draft' --vault 3b241101-e2bb-4255-8caf-4136c566a962
  mu notes search kubernetes --limit 3 --ndjson | jq -r .note.id | xargs -n1 mu notes get`

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
	addVaultFlags(cmd, &c.vault, false)
	addListFlags(cmd, &c.list)
	return cmd
}

func (c *notesSearchCmd) run(cmd *cobra.Command, text string) error {
	vault, page, err := c.request(text)
	if err != nil {
		return err
	}
	r, err := c.app.renderer(c.list.ndjson)
	if err != nil {
		return err
	}
	notes, err := c.app.notes(cmd.Context(), c.vault.vault)
	if err != nil {
		return err
	}
	body, err := notes.SearchNotes(cmd.Context(), text, vault, page)
	if err != nil {
		return err
	}
	return r.list(body, searchColumns, searchRow)
}

// request validates the query and the flags, and resolves the optional vault.
func (c *notesSearchCmd) request(text string) (string, client.Page, error) {
	if strings.TrimSpace(text) == "" {
		return "", client.Page{}, failure.New(failure.Usage, "QUERY", fmt.Sprintf("empty query %q, expected words to search", text))
	}
	page, err := c.list.page()
	if err != nil {
		return "", client.Page{}, err
	}
	resolved, _, err := c.app.resolve(c.vault.vault)
	return resolved.Vault.Value, page, err
}

func searchRow(raw json.RawMessage) ([]string, error) {
	var hit schema.SearchHit
	if err := json.Unmarshal(raw, &hit); err != nil {
		return nil, malformedRecord(raw, err)
	}
	return []string{hit.Note.ID, hit.Note.Path, output.Snippet(hit.Snippet)}, nil
}
