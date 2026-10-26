package cli

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
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

const notesListLong = `List the notes of a vault, by path, without their content. --public SLUG
lists a published vault instead, without credentials.

--orphan keeps the notes without any link, --community N those of one community
(ids from mu graph communities), and --sort orders by pagerank, degree or updated
time, descending unless --direction asc. Notes the indexer has not ranked yet
have no community and come last in those orders.

The listing is paginated: --limit is clamped to 100 and --offset to 10000 by the
server. Fetch a note's content with mu notes get, one call per note.

` + publicNeedsServer

type notesListCmd struct {
	app    *App
	vault  vaultFlags
	list   listFlags
	filter noteFilterFlags
}

func newNotesListCmd(app *App) *cobra.Command {
	c := &notesListCmd{app: app}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the notes of a vault, without content",
		Long:  notesListLong,
		Example: `  mu notes list --vault 3b241101-e2bb-4255-8caf-4136c566a962 --limit 20
  mu notes list --limit 100 --offset 100 --ndjson | jq -r .path
  mu notes list --sort pagerank --limit 10
  mu notes list --orphan --ndjson | jq -r .path
  mu notes list --public my-garden --limit 20`,
		Args: exactArgs(),
		RunE: func(cmd *cobra.Command, _ []string) error { return c.run(cmd) },
	}
	addVaultFlags(cmd, &c.vault, true)
	addListFlags(cmd, &c.list)
	addNoteFilterFlags(cmd, &c.filter)
	return cmd
}

func (c *notesListCmd) run(cmd *cobra.Command) error {
	filter, err := c.filter.noteFilter(cmd)
	if err != nil {
		return err
	}
	page, r, err := c.app.startListing(c.list)
	if err != nil {
		return err
	}
	body, err := c.fetch(cmd.Context(), filter, page)
	if err != nil {
		return err
	}
	return r.list(body, noteListColumns, noteListRow)
}

// fetch lists the vault of the account, or the public one with --public.
func (c *notesListCmd) fetch(ctx context.Context, filter client.NoteFilter, page client.Page) (io.ReadCloser, error) {
	svc, target, err := c.app.connectVault(ctx, c.vault)
	if err != nil {
		return nil, err
	}
	if target.PublicSlug != "" {
		public, err := publicOf(svc)
		if err != nil {
			return nil, err
		}
		return publicRead(public.PublicNotes(ctx, target.PublicSlug, filter, page))
	}
	notes, err := notesOf(svc)
	if err != nil {
		return nil, err
	}
	return notes.ListNotes(ctx, target.ID, filter, page)
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
for byte and unsanitized, to save it to a file and edit it back with mu notes edit.
--public SLUG reads the note from a published vault, without credentials.

` + publicNeedsServer

type notesGetCmd struct {
	app     *App
	content bool
	public  string
}

func newNotesGetCmd(app *App) *cobra.Command {
	c := &notesGetCmd{app: app}
	cmd := &cobra.Command{
		Use:   "get ID",
		Short: "Print a note with its metadata and content",
		Long:  notesGetLong,
		Example: `  mu notes get f47ac10b-58cc-4372-a567-0e02b2c3d479
  mu notes get f47ac10b-58cc-4372-a567-0e02b2c3d479 --content > meeting.md
  mu notes get f47ac10b-58cc-4372-a567-0e02b2c3d479 --json | jq -r .updated_at
  mu notes get f47ac10b-58cc-4372-a567-0e02b2c3d479 --public my-garden`,
		Args: exactArgs("ID"),
		RunE: func(cmd *cobra.Command, args []string) error { return c.run(cmd, args[0]) },
	}
	cmd.Flags().BoolVar(&c.content, "content", false, "print only the content, unchanged and unsanitized")
	addPublicFlag(cmd, &c.public)
	return cmd
}

func (c *notesGetCmd) run(cmd *cobra.Command, id string) error {
	if c.content && c.app.global.json {
		return failure.New(failure.Usage, "", "--content and --json are exclusive, expected at most one")
	}
	body, err := c.fetch(cmd.Context(), id)
	if err != nil {
		return err
	}
	return c.app.renderNote(body, c.print)
}

// fetch reads note id from the account, or from the public vault with --public.
func (c *notesGetCmd) fetch(ctx context.Context, id string) (io.ReadCloser, error) {
	if c.public != "" {
		public, err := c.app.publicReader(ctx)
		if err != nil {
			return nil, err
		}
		return publicRead(public.PublicNote(ctx, c.public, id))
	}
	notes, err := c.app.notes(ctx, "")
	if err != nil {
		return nil, err
	}
	return notes.GetNote(ctx, id)
}

func (c *notesGetCmd) print(p *output.Printer, note schema.Note) error {
	if c.content {
		return p.Raw(note.Content)
	}
	return printNote(p, note)
}

// renderNote prints a note body: unchanged with --json, decoded and handed to print
// otherwise.
func (a *App) renderNote(body io.ReadCloser, print func(*output.Printer, schema.Note) error) error {
	r, err := a.renderer(false)
	if err != nil {
		_ = body.Close()
		return err
	}
	return r.object(body, func(p *output.Printer, raw json.RawMessage) error {
		var note schema.Note
		if err := json.Unmarshal(raw, &note); err != nil {
			return malformedRecord(raw, err)
		}
		return print(p, note)
	})
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
		{Key: "degree", Value: dashedInt(note.Degree)},
		{Key: "pagerank", Value: optionalFloat(note.PageRank)},
		{Key: "community", Value: dashedInt64(note.Community)},
	}
}

// contentMark is the delimiter tag of a content block: a hash prefix of the content,
// which the content itself cannot contain.
func contentMark(content string) string {
	sum := sha256.Sum256([]byte(content))
	return "sha256:" + hex.EncodeToString(sum[:8])
}

func dashedInt(value *int) string {
	if value == nil {
		return "-"
	}
	return strconv.Itoa(*value)
}

func dashedInt64(value *int64) string {
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
