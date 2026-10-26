// Package cli holds the cobra command tree of mu. Cobra stays in this package: the
// commands reach the server through the ports in Services and print through the
// output package.
package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/output"
)

// commandConstructors builds the top level commands, in the order mu --help lists
// them. A new top level command registers itself with one line here.
var commandConstructors = []func(*App) *cobra.Command{
	newConfigCmd,
}

const rootExample = `  mu config set url https://markupp.example.com
  mu config show --json`

// Execute runs mu with args and returns the process exit code. Errors are printed on
// stderr; stdout only ever carries the data of a successful command.
func Execute(ctx context.Context, deps Deps, args []string) int {
	app := newApp(deps)
	root := newRoot(app)
	root.SetArgs(args)
	cmd, err := root.ExecuteContextC(ctx)
	if err == nil {
		return 0
	}
	err = app.classify(err)
	app.report(cmd, err)
	return failure.ExitCode(err)
}

func newRoot(app *App) *cobra.Command {
	root := &cobra.Command{
		Use:               "mu",
		Short:             "Command line client of the Markupp server",
		Long:              "mu reads and writes the notes of a Markupp server: text by default, the API payload with --json.",
		Example:           rootExample,
		Version:           app.deps.Version,
		SilenceErrors:     true,
		SilenceUsage:      true,
		PersistentPreRunE: app.prepare,
	}
	root.SetVersionTemplate("mu {{.Version}}\n")
	root.SetOut(app.deps.Stdout)
	root.SetErr(app.deps.Stderr)
	root.SetFlagErrorFunc(flagError)
	app.addGlobalFlags(root)
	for _, construct := range commandConstructors {
		root.AddCommand(construct(app))
	}
	attachSubcommands(app, root)
	return root
}

// classify turns the errors cobra raises before any command runs (unknown command,
// bad arguments) into usage failures. Errors of a running command keep their kind.
func (a *App) classify(err error) error {
	var classified *failure.Error
	if a.ran || errors.As(err, &classified) || errors.Is(err, context.Canceled) {
		return err
	}
	return failure.Wrap(failure.Usage, "", err)
}

// report prints the error, its hint and, for a usage error, where the help is.
func (a *App) report(cmd *cobra.Command, err error) {
	lines := []string{"Error: " + strings.TrimSpace(output.SanitizeBlock(err.Error()))}
	if hint := failure.HintOf(err); hint != "" {
		lines = append(lines, "Hint: "+output.SanitizeLine(hint))
	}
	if failure.KindOf(err) == failure.Usage && cmd != nil {
		lines = append(lines, fmt.Sprintf("Run '%s --help' for usage.", cmd.CommandPath()))
	}
	_, _ = fmt.Fprintln(a.deps.Stderr, strings.Join(lines, "\n"))
}
