package cli

import (
	"github.com/spf13/cobra"

	"github.com/markupp-labs/markupp-cli/internal/failure"
)

type notesSearchCmd struct {
	app   *App
	vault vaultFlags
	list  listFlags
}

func newNotesSearchCmd(app *App) *cobra.Command {
	c := &notesSearchCmd{app: app}
	cmd := &cobra.Command{
		Use:     "search QUERY",
		Short:   "Search notes by text",
		Example: "  mu notes search kubernetes --limit 5",
		Args:    exactArgs("QUERY"),
		RunE:    func(*cobra.Command, []string) error { return failure.New(failure.Internal, "", "not implemented") },
	}
	addVaultFlags(cmd, &c.vault, false)
	addListFlags(cmd, &c.list)
	return cmd
}
