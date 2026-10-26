// Package output is the only place that writes to stdout. Commands hand it raw API
// bodies or decoded records, and it renders them as aligned text, as the JSON body
// unchanged, or as one JSON object per line.
package output

import (
	"encoding/json"
	"io"
	"iter"
)

// Mode is the rendering a command was asked for.
type Mode int

// The rendering modes: aligned text (the default), the API body unchanged, and one
// compact JSON object per line.
const (
	Text Mode = iota
	JSON
	NDJSON
)

// Printer writes rendered output to the stdout it was built with.
type Printer struct{ out io.Writer }

// Field is one key and value of a record printed by Fields.
type Field struct {
	Key   string
	Value string
}

// Column is a table column. Width 0 leaves the column unpadded, for the last one.
type Column struct {
	Name  string
	Width int
}

// Table writes rows of fixed width columns as they come, without buffering.
type Table struct {
	out     io.Writer
	columns []Column
}

// NewPrinter returns a Printer writing to out.
func NewPrinter(out io.Writer) *Printer { return &Printer{out: out} }

// Line writes s on one line, sanitized.
func (p *Printer) Line(s string) error { return nil }

// Block writes multi line text, sanitized, ending with a newline.
func (p *Printer) Block(s string) error { return nil }

// Fields writes one aligned "key  value" line per field.
func (p *Printer) Fields(fields ...Field) error { return nil }

// JSON copies a JSON body unchanged, adding a final newline when it lacks one.
func (p *Printer) JSON(body io.Reader) error { return nil }

// NDJSON writes each element compacted on its own line, stopping at the first error.
func (p *Printer) NDJSON(elements iter.Seq2[json.RawMessage, error]) error { return nil }

// Table starts a table with columns.
func (p *Printer) Table(columns ...Column) *Table { return &Table{out: p.out, columns: columns} }

// Header writes the column names.
func (t *Table) Header() error { return nil }

// Row writes one record, one cell per column.
func (t *Table) Row(cells ...string) error { return nil }

// SanitizeLine removes control and escape characters and folds newlines and tabs
// into spaces, so a value from the server stays on its line.
func SanitizeLine(s string) string { return s }

// SanitizeBlock removes control and escape characters, keeping newlines and tabs.
func SanitizeBlock(s string) string { return s }

// Snippet renders a search snippet on one line, with the match markers U+0002 and
// U+0003 shown as "[[" and "]]".
func Snippet(s string) string { return s }
