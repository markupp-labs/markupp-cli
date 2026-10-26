package tui

import (
	"github.com/rivo/tview"
)

// record is one row of a listing.
type record struct {
	// id is what y copies and what Enter opens.
	id string
	// note is the note the note hotkeys act on, empty when the row is not a note.
	note string
	// label names the row in a confirmation, like the path of a note.
	label string
	cells []string
}

// listing describes a resource table: what it shows and what its keys do.
type listing struct {
	resource string
	context  string
	columns  []string
	load     loader
	// enter opens the selected row; nil means Enter does nothing.
	enter func(r record)
	// keys are the hotkeys of the rows, given the selected row.
	keys func(selected func() (record, bool)) []Hotkey
	// search runs /!term on the server; nil means the listing has no search.
	search func(term string)
}

// listView is the generic resource table: sortable, filterable and reloadable.
type listView struct {
	shell  *shell
	spec   listing
	table  *tview.Table
	all    []record
	shown  []record
	filter string
	order  sortOrder
}

func newListView(s *shell, spec listing) *listView { return *new(*listView) }

func (lv *listView) Crumb() string { return *new(string) }

// Title is resource(context)[count], with the filter when there is one.
func (lv *listView) Title() string { return *new(string) }

func (lv *listView) Widget() tview.Primitive { return *new(tview.Primitive) }

func (lv *listView) Refreshable() bool { return *new(bool) }

// Load reads the records and keeps the selection, filter and sort order.
func (lv *listView) Load() {}

func (lv *listView) setRecords(records []record) {}

func (lv *listView) Filter() string { return *new(string) }

func (lv *listView) SetFilter(term string) {}

// Search runs a server search when the listing has one.
func (lv *listView) Search(term string) {}

// Hotkeys are Enter, the keys of the listing and Shift plus a column initial to sort.
func (lv *listView) Hotkeys() []Hotkey { return *new([]Hotkey) }

func (lv *listView) open() {}

// selected is the row under the cursor; row 0 is the header.
func (lv *listView) selected() (record, bool) { return *new(record), *new(bool) }

func (lv *listView) sortKeys() []Hotkey { return *new([]Hotkey) }

func (lv *listView) sortBy(column int) {}

// apply filters and sorts the records, then draws them.
func (lv *listView) apply() {}

func (lv *listView) render() {}

func (lv *listView) renderHeader() {}

func (lv *listView) renderRow(row int, rec record) {}

// clampRow keeps the selection inside the rows, below the header.
func clampRow(row, count int) int { return *new(int) }

// filterRecords keeps the records with a cell containing term, ignoring case.
func filterRecords(records []record, term string) []record { return *new([]record) }

func matchesAny(cells []string, needle string) bool { return *new(bool) }
