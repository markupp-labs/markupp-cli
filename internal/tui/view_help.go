package tui

// newHelpView lists the keys of view, the global keys and the : commands.
func newHelpView(s *shell, view View) *listView { return *new(*listView) }

func keyRecords(scope string, keys []Hotkey) []record { return *new([]record) }

func commandRecords() []record { return *new([]record) }

func joinAliases(aliases []string) string { return *new(string) }
