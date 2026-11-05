// Package output is the only place that writes to stdout. Commands hand it raw API
// bodies or decoded records, and it renders them as aligned text, as the JSON body
// unchanged, or as one JSON object per line.
package output

import (
	"fmt"
	"io"
	"strings"
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

// NewPrinter returns a Printer writing to out.
func NewPrinter(out io.Writer) *Printer { return &Printer{out: out} }

// Line writes s on one line, sanitized.
func (p *Printer) Line(s string) error {
	return p.write(SanitizeLine(s) + "\n")
}

// Block writes multi line text, sanitized, ending with a newline.
func (p *Printer) Block(s string) error {
	text := SanitizeBlock(s)
	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	return p.write(text)
}

// Fields writes one aligned "key  value" line per field, the keys padded to the
// longest one.
func (p *Printer) Fields(fields ...Field) error {
	width := 0
	for _, field := range fields {
		width = max(width, len(field.Key))
	}
	for _, field := range fields {
		line := fmt.Sprintf("%-*s  %s", width, field.Key, SanitizeLine(field.Value))
		if err := p.write(line + "\n"); err != nil {
			return err
		}
	}
	return nil
}

func (p *Printer) write(text string) error {
	if _, err := io.WriteString(p.out, text); err != nil {
		return fmt.Errorf("writing to stdout: %w", err)
	}
	return nil
}
