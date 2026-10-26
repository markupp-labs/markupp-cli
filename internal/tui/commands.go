package tui

// command is a : command of the prompt.
type command struct {
	name    string
	aliases []string
	usage   string
	open    func(s *shell, arg string) error
}

// commandTable lists the : commands in the order of the help. A new resource adds
// its line here and its view in a file of its own.
func commandTable() []command { return *new([]command) }

// findCommand returns the command named word, by name or alias.
func findCommand(word string) (command, bool) { return *new(command), *new(bool) }

// commandNames are the names of the commands, for the error of an unknown one.
func commandNames() []string { return *new([]string) }

// completions are the command names that start with prefix.
func completions(prefix string) []string { return *new([]string) }
