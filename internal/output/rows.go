package output

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// Column is a table column. Width 0 leaves the column unpadded, for the last one.
type Column struct {
	Name  string
	Width int
}

// Table writes rows of fixed width columns as they come, without buffering: the widths
// are declared up front instead of measured over the whole result.
type Table struct {
	out     io.Writer
	columns []Column
}

const (
	columnGap   = "  "
	emptyCell   = "-"
	truncMarker = "..."
)

// Table starts a table with columns.
func (p *Printer) Table(columns ...Column) *Table {
	return &Table{out: p.out, columns: columns}
}

// Header writes the column names.
func (t *Table) Header() error {
	names := make([]string, len(t.columns))
	for i, column := range t.columns {
		names[i] = column.Name
	}
	return t.Row(names...)
}

// Row writes one record, one cell per column. A cell is sanitized onto one line,
// truncated to the column width, and an empty cell shows as "-".
func (t *Table) Row(cells ...string) error {
	if len(cells) != len(t.columns) {
		return fmt.Errorf("table row has %d cells %q, expected %d", len(cells), cells, len(t.columns))
	}
	parts := make([]string, len(cells))
	for i, cell := range cells {
		parts[i] = fitCell(cell, t.columns[i].Width)
	}
	line := strings.TrimRight(strings.Join(parts, columnGap), " ") + "\n"
	if _, err := io.WriteString(t.out, line); err != nil {
		return fmt.Errorf("writing to stdout: %w", err)
	}
	return nil
}

func fitCell(cell string, width int) string {
	text := SanitizeLine(cell)
	if text == "" {
		text = emptyCell
	}
	if width <= 0 {
		return text
	}
	return padRight(truncate(text, width), width)
}

func truncate(text string, width int) string {
	if utf8.RuneCountInString(text) <= width {
		return text
	}
	runes := []rune(text)
	if width <= len(truncMarker) {
		return string(runes[:width])
	}
	return string(runes[:width-len(truncMarker)]) + truncMarker
}

func padRight(text string, width int) string {
	missing := width - utf8.RuneCountInString(text)
	if missing <= 0 {
		return text
	}
	return text + strings.Repeat(" ", missing)
}
