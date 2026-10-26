package tui

// sortOrder is the sorted column of a table, -1 for the server order.
type sortOrder struct {
	column     int
	descending bool
}

// toggle sorts by column ascending, or flips the direction when it already sorts by it.
func (o sortOrder) toggle(column int) sortOrder { return *new(sortOrder) }

// arrow marks the header of the sorted column.
func (o sortOrder) arrow(column int) string { return *new(string) }

// sortRune is the Shift key that sorts by columns[i]: the upper case initial of the
// column, given only to the first column with that initial.
func sortRune(columns []string, i int) (rune, bool) { return *new(rune), *new(bool) }

func initialOf(column string) (rune, bool) { return *new(rune), *new(bool) }

// sortRecords sorts in place, stable, comparing numbers as numbers.
func sortRecords(records []record, order sortOrder) {}

func cellAt(rec record, column int) string { return *new(string) }

// compareCells compares two numbers numerically and anything else as text.
func compareCells(a, b string) int { return *new(int) }
