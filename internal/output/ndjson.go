package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"iter"
)

// NDJSON writes each element compacted on its own line as it arrives, stopping at the
// first error of the sequence.
func (p *Printer) NDJSON(elements iter.Seq2[json.RawMessage, error]) error {
	var line bytes.Buffer
	for element, err := range elements {
		if err != nil {
			return err
		}
		if err := p.compactLine(&line, element); err != nil {
			return err
		}
	}
	return nil
}

func (p *Printer) compactLine(line *bytes.Buffer, element json.RawMessage) error {
	line.Reset()
	if err := json.Compact(line, element); err != nil {
		return fmt.Errorf("element %.80s is not valid JSON, expected one JSON value: %w", element, err)
	}
	line.WriteByte('\n')
	if _, err := p.out.Write(line.Bytes()); err != nil {
		return fmt.Errorf("writing to stdout: %w", err)
	}
	return nil
}
