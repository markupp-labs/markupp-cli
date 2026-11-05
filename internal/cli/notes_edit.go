package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/output"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

// forcedSince is the lastModifiedAt sent with --force and no --since: the server
// ignores the version of a forced write, but the field is required.
const forcedSince = "0001-01-01T00:00:00Z"

const notesEditLong = `Replace the content of a note with --file, a path or - for stdin. mu never reads
stdin implicitly.

--since is the updated_at of the note as you last read it (mu notes get ID --json |
jq -r .updated_at). If the note changed on the server after that, the server
refuses the write with a conflict: mu exits 6 and never retries. Read the note
again, merge, and edit with the new updated_at.

Without --path the note keeps its path, which mu reads first with one more call.
On success mu prints the id, path and new updated_at of the note.`

const forceHelp = "overwrite even if the note changed on the server since --since: " +
	"the edits made there since then are lost"

type notesEditCmd struct {
	app   *App
	file  string
	since string
	path  string
	force bool
}

func newNotesEditCmd(app *App) *cobra.Command {
	c := &notesEditCmd{app: app}
	cmd := &cobra.Command{
		Use:   "edit ID",
		Short: "Replace the content of a note, refusing to overwrite newer edits",
		Long:  notesEditLong,
		Example: `  mu notes edit f47ac10b-58cc-4372-a567-0e02b2c3d479 --file meeting.md --since 2026-05-15T12:00:00.5Z
  mu notes get f47ac10b-58cc-4372-a567-0e02b2c3d479 --content | sed s/draft/final/ | mu notes edit f47ac10b-58cc-4372-a567-0e02b2c3d479 --file - --since 2026-05-15T12:00:00.5Z`,
		Args: exactArgs("ID"),
		RunE: func(cmd *cobra.Command, args []string) error { return c.run(cmd, args[0]) },
	}
	addFileFlag(cmd, &c.file)
	cmd.Flags().StringVar(&c.since, "since", "", "updated_at of the note as last read, RFC 3339 (required unless --force)")
	cmd.Flags().StringVar(&c.path, "path", "", "new path of the note (default: keep the current one)")
	cmd.Flags().BoolVar(&c.force, "force", false, forceHelp)
	return cmd
}

func (c *notesEditCmd) run(cmd *cobra.Command, id string) error {
	update, err := c.request()
	if err != nil {
		return err
	}
	notes, err := c.app.notes(cmd.Context(), "")
	if err != nil {
		return err
	}
	if update.Path, err = c.notePath(cmd.Context(), notes, id); err != nil {
		return err
	}
	body, err := notes.UpdateNote(cmd.Context(), id, update)
	if err != nil {
		return conflictHint(err, id)
	}
	return c.app.renderNote(body, printEdited)
}

// request validates --since and reads the content, before anything reaches the server.
func (c *notesEditCmd) request() (schema.NoteUpdateRequest, error) {
	since, err := c.version()
	if err != nil {
		return schema.NoteUpdateRequest{}, err
	}
	content, err := c.app.readContent(c.file)
	if err != nil {
		return schema.NoteUpdateRequest{}, err
	}
	return schema.NoteUpdateRequest{Content: content, LastModifiedAt: since, Force: c.force}, nil
}

// version returns the lastModifiedAt to send: --since, or the zero time when --force
// stands alone.
func (c *notesEditCmd) version() (string, error) {
	if c.since == "" && c.force {
		return forcedSince, nil
	}
	if c.since == "" {
		return "", failure.New(failure.Usage, "--since", "missing, expected the updated_at of the note as last read, or --force to overwrite")
	}
	if _, err := time.Parse(time.RFC3339Nano, c.since); err != nil {
		return "", failure.New(failure.Usage, "--since", fmt.Sprintf("invalid time %q, expected RFC 3339 like 2026-05-15T12:00:00Z", c.since)).WithCause(err)
	}
	return c.since, nil
}

// notePath is --path, or the current path of the note read from the server.
func (c *notesEditCmd) notePath(ctx context.Context, notes NoteReader, id string) (string, error) {
	if c.path != "" {
		return c.path, nil
	}
	body, err := notes.GetNote(ctx, id)
	if err != nil {
		return "", err
	}
	raw, err := readObject(body)
	if err != nil {
		return "", err
	}
	return decodeNotePath(raw)
}

// conflictHint explains a version conflict; other errors pass unchanged.
func conflictHint(err error, id string) error {
	var classified *failure.Error
	if !errors.As(err, &classified) || classified.Kind != failure.Conflict || classified.Code != "conflict" {
		return err
	}
	return classified.WithHint(fmt.Sprintf("the note changed since --since; run mu notes get %s, merge, and retry with its updated_at, or pass --force to discard the server edits", id))
}

func printEdited(p *output.Printer, note schema.Note) error {
	return p.Fields(
		output.Field{Key: "id", Value: note.ID},
		output.Field{Key: "path", Value: note.Path},
		output.Field{Key: "updated_at", Value: note.UpdatedAt},
	)
}

// decodeNotePath returns the path of a note body.
func decodeNotePath(raw json.RawMessage) (string, error) {
	note, err := decodeRecord[schema.Note](raw)
	if err != nil {
		return "", err
	}
	return note.Path, nil
}
