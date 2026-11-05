package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

func countNamed(root *cobra.Command, name string) int {
	count := 0
	for _, child := range root.Commands() {
		if child.Name() == name {
			count++
		}
	}
	return count
}

func TestSubcommandsJoinAGroupAnotherFileBuilt(t *testing.T) {
	app := newApp(newTestMu().deps(nil, nil))
	root := &cobra.Command{Use: "mu"}
	owned := newGroup("notes", "Owned elsewhere", "  mu notes list")
	root.AddCommand(owned)
	attachSubcommands(app, root)
	if got := countNamed(root, "notes"); got != 1 {
		t.Fatalf("root has %d notes commands, want 1", got)
	}
	if len(owned.Commands()) != len(subcommandConstructors) {
		t.Errorf("the existing notes group got %d subcommands, want %d", len(owned.Commands()), len(subcommandConstructors))
	}
}

func TestSubcommandsCreateTheirGroupWhenMissing(t *testing.T) {
	app := newApp(newTestMu().deps(nil, nil))
	root := &cobra.Command{Use: "mu"}
	attachSubcommands(app, root)
	if got := countNamed(root, "notes"); got != 1 {
		t.Fatalf("root has %d notes commands, want 1 created from the fallback", got)
	}
}
