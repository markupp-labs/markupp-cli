package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/output"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// updatedWidth fits an RFC 3339 time with nanoseconds in UTC, so the value printed is
// the exact one notes edit --since expects.
const updatedWidth = 30

var noteListColumns = []output.Column{{Name: "ID", Width: 36}, {Name: "UPDATED", Width: updatedWidth}, {Name: "PATH"}}

const notesListLong = `List the notes of a vault, by path, without their content.

The listing is paginated: --limit is clamped to 100 and --offset to 10000 by the
server. Fetch a note's content with mu notes get, one call per note.`

type notesListCmd struct {
	app   *App
	vault vaultFlags
	list  listFlags
}

func newNotesListCmd(app *App) *cobra.Command {
	c := &notesListCmd{app: app}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the notes of a vault, without content",
		Long:  notesListLong,
		Example: `  mu notes list --vault 3b241101-e2bb-4255-8caf-4136c566a962 --limit 20
  mu notes list --limit 100 --offset 100 --ndjson | jq -r .path`,
		Args: exactArgs(),
		RunE: func(cmd *cobra.Command, _ []string) error { return c.run(cmd) },
	}
	addVaultFlags(cmd, &c.vault, false)
	addListFlags(cmd, &c.list)
	return cmd
}

func (c *notesListCmd) run(cmd *cobra.Command) error {
	target, page, err := c.request()
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
	body, err := notes.ListNotes(cmd.Context(), target.ID, page)
	if err != nil {
		return err
	}
	return r.list(body, noteListColumns, noteListRow)
}

// request validates the flags before anything reaches the server.
func (c *notesListCmd) request() (vaultTarget, client.Page, error) {
	target, err := c.app.resolveVault(c.vault)
	if err != nil {
		return vaultTarget{}, client.Page{}, err
	}
	page, err := c.list.page()
	return target, page, err
}

func noteListRow(raw json.RawMessage) ([]string, error) {
	var note schema.NoteListItem
	if err := json.Unmarshal(raw, &note); err != nil {
		return nil, malformedRecord(raw, err)
	}
	return []string{note.ID, note.UpdatedAt, note.Path}, nil
}

const notesGetLong = `Print a note: its metadata, then its content between two delimiter lines that
carry a hash of the content, so the content cannot fake the end of the block.

The content is sanitized for the terminal. --content prints only the content, byte
for byte and unsanitized, to save it to a file and edit it back with mu notes edit.`

type notesGetCmd struct {
	app     *App
	content bool
}

func newNotesGetCmd(app *App) *cobra.Command {
	c := &notesGetCmd{app: app}
	cmd := &cobra.Command{
		Use:   "get ID",
		Short: "Print a note with its metadata and content",
		Long:  notesGetLong,
		Example: `  mu notes get f47ac10b-58cc-4372-a567-0e02b2c3d479
  mu notes get f47ac10b-58cc-4372-a567-0e02b2c3d479 --content > meeting.md
  mu notes get f47ac10b-58cc-4372-a567-0e02b2c3d479 --json | jq -r .updated_at`,
		Args: exactArgs("ID"),
		RunE: func(cmd *cobra.Command, args []string) error { return c.run(cmd, args[0]) },
	}
	cmd.Flags().BoolVar(&c.content, "content", false, "print only the content, unchanged and unsanitized")
	return cmd
}

func (c *notesGetCmd) run(cmd *cobra.Command, id string) error {
	if c.content && c.app.global.json {
		return failure.New(failure.Usage, "", "--content and --json are exclusive, expected at most one")
	}
	r, err := c.app.renderer(false)
	if err != nil {
		return err
	}
	notes, err := c.app.notes(cmd.Context(), "")
	if err != nil {
		return err
	}
	body, err := notes.GetNote(cmd.Context(), id)
	if err != nil {
		return err
	}
	return r.object(body, c.render)
}

func (c *notesGetCmd) render(p *output.Printer, raw json.RawMessage) error {
	var note schema.Note
	if err := json.Unmarshal(raw, &note); err != nil {
		return malformedRecord(raw, err)
	}
	if c.content {
		return p.Raw(note.Content)
	}
	return printNote(p, note)
}

// printNote prints the metadata of note, then its content between delimiter lines.
func printNote(p *output.Printer, note schema.Note) error {
	if err := p.Fields(noteFields(note)...); err != nil {
		return err
	}
	mark := contentMark(note.Content)
	if err := p.Line("--- content " + mark + " ---"); err != nil {
		return err
	}
	if err := p.Block(note.Content); err != nil {
		return err
	}
	return p.Line("--- end of content " + mark + " ---")
}

func noteFields(note schema.Note) []output.Field {
	return []output.Field{
		{Key: "id", Value: note.ID},
		{Key: "path", Value: note.Path},
		{Key: "vault_id", Value: note.VaultID},
		{Key: "created_at", Value: note.CreatedAt},
		{Key: "updated_at", Value: note.UpdatedAt},
		{Key: "content_sha256", Value: note.ContentSHA256},
		{Key: "degree", Value: optionalInt(note.Degree)},
		{Key: "pagerank", Value: optionalFloat(note.PageRank)},
		{Key: "community", Value: optionalInt64(note.Community)},
	}
}

// contentMark is the delimiter tag of a content block: a hash prefix of the content,
// which the content itself cannot contain.
func contentMark(content string) string {
	sum := sha256.Sum256([]byte(content))
	return "sha256:" + hex.EncodeToString(sum[:8])
}

func optionalInt(value *int) string {
	if value == nil {
		return "-"
	}
	return strconv.Itoa(*value)
}

func optionalInt64(value *int64) string {
	if value == nil {
		return "-"
	}
	return strconv.FormatInt(*value, 10)
}

func optionalFloat(value *float64) string {
	if value == nil {
		return "-"
	}
	return strconv.FormatFloat(*value, 'g', -1, 64)
}
