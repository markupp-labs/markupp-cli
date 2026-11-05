package tui

import (
	"slices"
	"strings"
)

// command is a : command of the prompt.
type command struct {
	name    string
	aliases []string
	usage   string
	open    func(s *shell, arg string) error
}

// commandTable lists the : commands in the order of the help. A new resource adds
// its line here and its view in a file of its own.
func commandTable() []command {
	return []command{
		{name: "notes", aliases: []string{"n", "no"}, usage: ":notes [vault id]", open: openNotes},
		{name: "note", aliases: []string{"nt"}, usage: ":note <note id>", open: openNote},
		{name: "graph", aliases: []string{"g", "xray"}, usage: ":graph <note id>", open: openGraph},
		{name: "communities", aliases: []string{"cm", "com"}, usage: ":communities [vault id]", open: openCommunities},
		{name: "links", aliases: []string{"ln"}, usage: ":links [vault id]", open: openVaultLinks},
		{name: "vaults", aliases: []string{"v", "va"}, usage: ":vaults", open: openVaults},
		{name: "members", aliases: []string{"mb"}, usage: ":members", open: openMembers},
		{name: "invites", aliases: []string{"inv"}, usage: ":invites", open: openInvites},
		{name: "tenants", aliases: []string{"tn", "te"}, usage: ":tenants", open: openTenants},
		{name: "notifications", aliases: []string{"notif", "bell"}, usage: ":notifications", open: openNotifications},
		{name: "public", aliases: []string{"pub"}, usage: ":public <slug>", open: openPublic},
	}
}

// findCommand returns the command named word, by name or alias.
func findCommand(word string) (command, bool) {
	for _, cmd := range commandTable() {
		if cmd.name == word || slices.Contains(cmd.aliases, word) {
			return cmd, true
		}
	}
	return command{}, false
}

// commandNames are the names of the commands, for the error of an unknown one.
func commandNames() []string {
	var names []string
	for _, cmd := range commandTable() {
		names = append(names, cmd.name)
	}
	return append(names, "q")
}

// completions are the command names that start with prefix.
func completions(prefix string) []string {
	prefix = strings.ToLower(prefix)
	var names []string
	for _, name := range commandNames() {
		if strings.HasPrefix(name, prefix) {
			names = append(names, name)
		}
	}
	return names
}
