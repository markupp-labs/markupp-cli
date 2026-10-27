package tui

import "context"

// newHelpView lists the keys of view, the global keys and the : commands.
func newHelpView(s *shell, view View) *listView {
	rows := keyRecords("view", view.Hotkeys())
	rows = append(rows, keyRecords("global", s.globalHotkeys())...)
	rows = append(rows, commandRecords()...)
	return newListView(s, listing{
		resource: "help",
		context:  view.Crumb(),
		columns:  []string{"SCOPE", "KEY", "ACTION"},
		load:     func(context.Context) ([]record, error) { return rows, nil },
	})
}

func keyRecords(scope string, keys []Hotkey) []record {
	rows := make([]record, 0, len(keys))
	for _, key := range keys {
		rows = append(rows, record{cells: []string{scope, key.Name(), key.Label}})
	}
	return rows
}

func commandRecords() []record {
	var rows []record
	for _, cmd := range commandTable() {
		rows = append(rows, record{cells: []string{"command", cmd.usage, "aliases: " + joinAliases(cmd.aliases)}})
	}
	return append(rows, record{cells: []string{"command", ":q", "quit"}})
}

func joinAliases(aliases []string) string {
	text := ""
	for i, alias := range aliases {
		if i > 0 {
			text += ", "
		}
		text += ":" + alias
	}
	return text
}
