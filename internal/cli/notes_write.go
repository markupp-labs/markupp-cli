package cli

import (
	"github.com/spf13/cobra"

	"github.com/markupp-labs/markupp-cli/internal/failure"
)

var errWritePending = failure.New(failure.Internal, "", "not implemented")

func newNotesCreateCmd(*App) *cobra.Command {
	cmd := &cobra.Command{Use: "create PATH", Short: "Create a note", Example: "  mu notes create a.md --file a.md",
		RunE: func(*cobra.Command, []string) error { return errWritePending }}
	cmd.Flags().String("file", "", "")
	cmd.Flags().String("vault", "", "")
	return cmd
}

func newNotesEditCmd(*App) *cobra.Command {
	cmd := &cobra.Command{Use: "edit ID", Short: "Edit a note", Example: "  mu notes edit n1 --file a.md",
		RunE: func(*cobra.Command, []string) error { return errWritePending }}
	cmd.Flags().String("file", "", "")
	cmd.Flags().String("since", "", "")
	cmd.Flags().String("path", "", "")
	cmd.Flags().Bool("force", false, "")
	return cmd
}

func newNotesDeleteCmd(*App) *cobra.Command {
	cmd := &cobra.Command{Use: "delete ID", Short: "Delete a note", Example: "  mu notes delete n1 --yes",
		RunE: func(*cobra.Command, []string) error { return errWritePending }}
	cmd.Flags().Bool("yes", false, "")
	return cmd
}

func (a *App) readContent(string) (string, error) { return "", errWritePending }
