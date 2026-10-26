package tui

import (
	"time"

	"github.com/rivo/tview"
)

// headerHeight is the number of lines of the header, and of hotkeys per column.
const headerHeight = 6

// header shows the session on the left and the hotkeys of the view on the right.
type header struct {
	widget *tview.Flex
	info   *tview.TextView
	keys   *tview.TextView
}

func newHeader() *header { return *new(*header) }

// headerState is everything the header draws.
type headerState struct {
	session Session
	tenant  string
	account string
	vault   string
	refresh string
	now     time.Time
	keys    []Hotkey
}

func (s *shell) headerState(top View) headerState { return *new(headerState) }

func (h *header) render(state headerState) {}

// infoText is the left side of the header, one setting per line.
func infoText(state headerState) string { return *new(string) }

func orDash(value string) string { return *new(string) }

// tokenText says where the token comes from and when a stored login expires.
func tokenText(session Session, now time.Time) string { return *new(string) }

// hotkeyColumns lays the hotkeys out in columns of rows lines.
func hotkeyColumns(keys []Hotkey, rows int) string { return *new(string) }

// majorKeys leaves out the keys that show only in ? help.
func majorKeys(keys []Hotkey) []Hotkey { return *new([]Hotkey) }

// refreshText describes the auto refresh for the header.
func (s *shell) refreshText() string { return *new(string) }
