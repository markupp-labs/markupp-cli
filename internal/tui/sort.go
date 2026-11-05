package tui

import (
	"cmp"
	"slices"
	"strconv"
	"unicode"
)

// sortOrder is the sorted column of a table, -1 for the server order.
type sortOrder struct {
	column     int
	descending bool
}

// toggle sorts by column ascending, or flips the direction when it already sorts by it.
func (o sortOrder) toggle(column int) sortOrder {
	if o.column == column {
		return sortOrder{column: column, descending: !o.descending}
	}
	return sortOrder{column: column}
}

// arrow marks the header of the sorted column.
func (o sortOrder) arrow(column int) string {
	switch {
	case o.column != column:
		return ""
	case o.descending:
		return " v"
	default:
		return " ^"
	}
}

// sortRune is the Shift key that sorts by columns[i]: the upper case initial of the
// column, given only to the first column with that initial.
func sortRune(columns []string, i int) (rune, bool) {
	initial, ok := initialOf(columns[i])
	if !ok {
		return 0, false
	}
	for _, earlier := range columns[:i] {
		if other, _ := initialOf(earlier); other == initial {
			return 0, false
		}
	}
	return initial, true
}

func initialOf(column string) (rune, bool) {
	for _, r := range column {
		return unicode.ToUpper(r), unicode.IsLetter(r)
	}
	return 0, false
}

// sortRecords sorts in place, stable, comparing numbers as numbers.
func sortRecords(records []record, order sortOrder) {
	if order.column < 0 {
		return
	}
	slices.SortStableFunc(records, func(a, b record) int {
		result := compareCells(cellAt(a, order.column), cellAt(b, order.column))
		if order.descending {
			return -result
		}
		return result
	})
}

func cellAt(rec record, column int) string {
	if column >= len(rec.cells) {
		return ""
	}
	return rec.cells[column]
}

// compareCells compares two numbers numerically and anything else as text.
func compareCells(a, b string) int {
	x, errA := strconv.ParseFloat(a, 64)
	y, errB := strconv.ParseFloat(b, 64)
	if errA == nil && errB == nil {
		return cmp.Compare(x, y)
	}
	return cmp.Compare(a, b)
}
