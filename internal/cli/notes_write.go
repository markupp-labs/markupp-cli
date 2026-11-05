package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/markupp-labs/markupp-cli/internal/output"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

const notesCreateLong = `Create a note at PATH in a vault, with the content of --file, and print its id.

PATH is unique in the vault, and its last segment without .md is the note title.
The content comes from --file, a path or - for stdin; mu never reads stdin
implicitly. With --json the created note is printed instead of the id.`

type notesCreateCmd struct {
	app   *App
	vault vaultFlags
	file  string
}

func newNotesCreateCmd(app *App) *cobra.Command {
	c := &notesCreateCmd{app: app}
	cmd := &cobra.Command{
		Use:   "create PATH",
		Short: "Create a note and print its id",
		Long:  notesCreateLong,
		Example: `  mu notes create notas/meeting.md --file meeting.md
  echo '# Ideas' | mu notes create ideas.md --file - --vault 3b241101-e2bb-4255-8caf-4136c566a962`,
		Args: exactArgs("PATH"),
		RunE: func(cmd *cobra.Command, args []string) error { return c.run(cmd, args[0]) },
	}
	addVaultFlags(cmd, &c.vault, false)
	addFileFlag(cmd, &c.file)
	return cmd
}

func (c *notesCreateCmd) run(cmd *cobra.Command, path string) error {
	target, err := c.app.resolveVault(c.vault)
	if err != nil {
		return err
	}
	content, err := c.app.readContent(c.file)
	if err != nil {
		return err
	}
	notes, err := c.app.notes(cmd.Context(), c.vault.vault)
	if err != nil {
		return err
	}
	body, err := notes.CreateNote(cmd.Context(), target.ID, schema.NoteCreateRequest{Path: path, Content: content})
	if err != nil {
		return err
	}
	return c.app.renderNote(body, printNoteID)
}

func printNoteID(p *output.Printer, note schema.Note) error { return p.Line(note.ID) }

const notesDeleteLong = `Delete a note. The deletion cannot be undone, so it requires --yes: without it
nothing is sent. The confirmation goes to stderr; stdout stays empty.`

type notesDeleteCmd struct {
	app *App
	yes bool
}

func newNotesDeleteCmd(app *App) *cobra.Command {
	c := &notesDeleteCmd{app: app}
	cmd := &cobra.Command{
		Use:     "delete ID",
		Short:   "Delete a note (requires --yes)",
		Long:    notesDeleteLong,
		Example: "  mu notes delete f47ac10b-58cc-4372-a567-0e02b2c3d479 --yes",
		Args:    exactArgs("ID"),
		RunE:    func(cmd *cobra.Command, args []string) error { return c.run(cmd, args[0]) },
	}
	addYesFlag(cmd, &c.yes, "the note and its content are deleted for good")
	return cmd
}

func (c *notesDeleteCmd) run(cmd *cobra.Command, id string) error {
	if err := requireYes(c.yes, "delete note", id); err != nil {
		return err
	}
	notes, err := c.app.notes(cmd.Context(), "")
	if err != nil {
		return err
	}
	if err := notes.DeleteNote(cmd.Context(), id); err != nil {
		return err
	}
	_, err = fmt.Fprintf(c.app.deps.Stderr, "deleted note %s\n", output.SanitizeLine(id))
	return err
}
