package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
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

func newListView(s *shell, spec listing) *listView {
	table := tview.NewTable().SetSelectable(true, false).SetFixed(1, 0)
	lv := &listView{shell: s, spec: spec, table: table, order: sortOrder{column: -1}}
	lv.render()
	return lv
}

func (lv *listView) Crumb() string { return lv.spec.resource }

// Title is resource(context)[count], with the filter when there is one.
func (lv *listView) Title() string {
	title := lv.spec.resource
	if lv.spec.context != "" {
		title += "(" + lv.spec.context + ")"
	}
	title += "[" + strconv.Itoa(len(lv.shown)) + "]"
	if lv.filter != "" {
		title += " /" + lv.filter
	}
	return title
}

func (lv *listView) Widget() tview.Primitive { return lv.table }

func (lv *listView) Refreshable() bool { return true }

// Load reads the records and keeps the selection, filter and sort order.
func (lv *listView) Load() {
	lv.shell.async(func(ctx context.Context) (func(), error) {
		records, err := lv.spec.load(ctx)
		return func() { lv.setRecords(records) }, err
	})
}

func (lv *listView) setRecords(records []record) {
	lv.all = records
	lv.apply()
}

func (lv *listView) Filter() string { return lv.filter }

func (lv *listView) SetFilter(term string) {
	lv.filter = term
	lv.apply()
}

// Search runs a server search when the listing has one.
func (lv *listView) Search(term string) {
	if lv.spec.search == nil {
		lv.shell.flashError(fmt.Errorf("%s has no server search, expected / with plain text", lv.spec.resource))
		return
	}
	lv.spec.search(term)
}

// Hotkeys are Enter, the keys of the listing and Shift plus a column initial to sort.
func (lv *listView) Hotkeys() []Hotkey {
	var keys []Hotkey
	if lv.spec.enter != nil {
		keys = append(keys, ctrlKey(tcell.KeyEnter, "open", lv.open))
	}
	if lv.spec.keys != nil {
		keys = append(keys, lv.spec.keys(lv.selected)...)
	}
	return append(keys, lv.sortKeys()...)
}

func (lv *listView) open() {
	if row, ok := lv.selected(); ok {
		lv.spec.enter(row)
	}
}

// selected is the row under the cursor; row 0 is the header.
func (lv *listView) selected() (record, bool) {
	row, _ := lv.table.GetSelection()
	if row < 1 || row > len(lv.shown) {
		return record{}, false
	}
	return lv.shown[row-1], true
}

func (lv *listView) sortKeys() []Hotkey {
	var keys []Hotkey
	for i, column := range lv.spec.columns {
		initial, ok := sortRune(lv.spec.columns, i)
		if !ok {
			continue
		}
		index := i
		key := runeKey(initial, "sort "+column, func() { lv.sortBy(index) })
		key.minor = true
		keys = append(keys, key)
	}
	return keys
}

func (lv *listView) sortBy(column int) {
	lv.order = lv.order.toggle(column)
	lv.apply()
}

// apply filters and sorts the records, then draws them.
func (lv *listView) apply() {
	lv.shown = filterRecords(lv.all, lv.filter)
	sortRecords(lv.shown, lv.order)
	lv.render()
	lv.shell.refreshChrome()
}

func (lv *listView) render() {
	row, _ := lv.table.GetSelection()
	lv.table.Clear()
	lv.renderHeader()
	for i, rec := range lv.shown {
		lv.renderRow(i+1, rec)
	}
	// Scrolling to the top first drops the tracking of the end that tview turns on
	// while the table is empty; Select then scrolls back to the selection.
	lv.table.ScrollToBeginning().Select(clampRow(row, len(lv.shown)), 0)
}

func (lv *listView) renderHeader() {
	for i, column := range lv.spec.columns {
		title := column + lv.order.arrow(i)
		cell := tview.NewTableCell(tview.Escape(title)).SetSelectable(false).SetExpansion(1)
		lv.table.SetCell(0, i, cell.SetAttributes(tcell.AttrBold))
	}
}

func (lv *listView) renderRow(row int, rec record) {
	for i, value := range rec.cells {
		lv.table.SetCell(row, i, tview.NewTableCell(plain(value)).SetExpansion(1).SetMaxWidth(60))
	}
}

// clampRow keeps the selection inside the rows, below the header.
func clampRow(row, count int) int {
	if row < 1 {
		return 1
	}
	if row > count && count > 0 {
		return count
	}
	return row
}

// filterRecords keeps the records with a cell containing term, ignoring case.
func filterRecords(records []record, term string) []record {
	needle := strings.ToLower(term)
	kept := make([]record, 0, len(records))
	for _, rec := range records {
		if needle == "" || matchesAny(rec.cells, needle) {
			kept = append(kept, rec)
		}
	}
	return kept
}

func matchesAny(cells []string, needle string) bool {
	for _, cell := range cells {
		if strings.Contains(strings.ToLower(cell), needle) {
			return true
		}
	}
	return false
}
