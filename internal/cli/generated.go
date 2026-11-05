package cli

import (
	"strings"

	"github.com/spf13/cobra"
)

// generatedExamples are the examples of the commands cobra generates, keyed by their
// path under mu: every command carries a runnable example, these included.
var generatedExamples = map[string]string{
	"help":                  "  mu help notes edit",
	"completion":            "  mu completion bash > ~/.local/share/bash-completion/completions/mu",
	"completion bash":       "  mu completion bash > ~/.local/share/bash-completion/completions/mu",
	"completion zsh":        "  mu completion zsh > \"${fpath[1]}/_mu\"",
	"completion fish":       "  mu completion fish > ~/.config/fish/completions/mu.fish",
	"completion powershell": "  mu completion powershell | Out-String | Invoke-Expression",
}

// addGeneratedCommands builds the help and completion commands now, instead of when
// cobra executes, and gives each one its example.
func addGeneratedCommands(root *cobra.Command) {
	root.InitDefaultHelpCmd()
	root.InitDefaultCompletionCmd()
	for path, example := range generatedExamples {
		if cmd, _, err := root.Find(strings.Fields(path)); err == nil && cmd != root {
			cmd.Example = example
		}
	}
}
