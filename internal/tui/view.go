package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// View is one screen of the view stack. Adding a resource means a new View and a line
// in commandTable; the shell does not change.
type View interface {
	// Crumb is the label of the view in the breadcrumb bar.
	Crumb() string
	// Title is the title of the main frame, resource(context)[count].
	Title() string
	// Widget is what the main frame shows.
	Widget() tview.Primitive
	// Hotkeys are the keys of the view, shown in the header and in ? help.
	Hotkeys() []Hotkey
	// Load reads the view from the server. A view without remote state does nothing.
	Load()
}

// Filterable is a view that / filters locally.
type Filterable interface {
	Filter() string
	SetFilter(term string)
}

// Searchable is a view where /!term runs a server search.
type Searchable interface {
	Search(term string)
}

// Refreshable is a view the auto refresh reloads.
type Refreshable interface {
	Refreshable() bool
}

// Hotkey is a key of a view.
type Hotkey struct {
	Key   tcell.Key
	Rune  rune
	Label string
	Run   func()
	// minor keys, like the sort keys, show only in ? help, not in the header.
	minor bool
}

// ctrlDelete deletes, as in K9s.
const ctrlDelete = tcell.KeyCtrlD

// runeKey is a hotkey on a printable key.
func runeKey(r rune, label string, run func()) Hotkey {
	return Hotkey{Key: tcell.KeyRune, Rune: r, Label: label, Run: run}
}

// ctrlKey is a hotkey on a control key.
func ctrlKey(key tcell.Key, label string, run func()) Hotkey {
	return Hotkey{Key: key, Label: label, Run: run}
}

// matches reports whether event is the key of h.
func (h Hotkey) matches(event *tcell.EventKey) bool {
	if h.Key != event.Key() {
		return false
	}
	return h.Key != tcell.KeyRune || h.Rune == event.Rune()
}

// Name is how the header shows the key: <e>, <ctrl-d>, <enter>.
func (h Hotkey) Name() string {
	if h.Key == tcell.KeyRune {
		return "<" + string(h.Rune) + ">"
	}
	return "<" + keyName(h.Key) + ">"
}

var keyNames = map[tcell.Key]string{
	tcell.KeyEnter: "enter", tcell.KeyEscape: "esc", tcell.KeyCtrlD: "ctrl-d",
	tcell.KeyCtrlR: "ctrl-r", tcell.KeyCtrlS: "ctrl-s", tcell.KeyTab: "tab",
}

func keyName(key tcell.Key) string {
	if name, ok := keyNames[key]; ok {
		return name
	}
	return tcell.KeyNames[key]
}

// findHotkey returns the hotkey of event among keys.
func findHotkey(keys []Hotkey, event *tcell.EventKey) (Hotkey, bool) {
	for _, key := range keys {
		if key.matches(event) {
			return key, true
		}
	}
	return Hotkey{}, false
}
