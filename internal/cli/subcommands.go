package cli

import "github.com/spf13/cobra"

// subcommand is a command that lives under a top level group another file may own,
// so two files can add commands to the same group without editing each other.
type subcommand struct {
	parent string
	build  func(*App) *cobra.Command
}

// subcommandConstructors attach commands under a group, in this order. When no
// constructor in commandConstructors builds the group, the fallback in
// fallbackGroups creates it.
var subcommandConstructors = []subcommand{}

// fallbackGroups build a group that subcommands need when nothing else built it.
var fallbackGroups = map[string]func() *cobra.Command{
	"notes": func() *cobra.Command {
		return newGroup("notes", "Read and write notes", "  mu notes links 3f2a9c1e-5b7d-4e8a-9c0f-1a2b3c4d5e6f")
	},
}

// attachSubcommands adds every subcommand to its group under root.
func attachSubcommands(app *App, root *cobra.Command) {
	for _, sub := range subcommandConstructors {
		groupOf(root, sub.parent).AddCommand(sub.build(app))
	}
}

// groupOf returns the top level command named name, creating it from fallbackGroups
// when it does not exist yet.
func groupOf(root *cobra.Command, name string) *cobra.Command {
	for _, child := range root.Commands() {
		if child.Name() == name {
			return child
		}
	}
	group := fallbackGroups[name]()
	root.AddCommand(group)
	return group
}
