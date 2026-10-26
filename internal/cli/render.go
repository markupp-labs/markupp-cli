package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"iter"

	"github.com/markupp-labs/markupp-cli/internal/client"
	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/output"
)

// renderer prints an API body in the mode the flags asked for.
type renderer struct {
	mode    output.Mode
	printer *output.Printer
}

// rowFunc turns one element of a listing into the cells of its text row.
type rowFunc func(raw json.RawMessage) ([]string, error)

// objectFunc prints one object as text.
type objectFunc func(p *output.Printer, raw json.RawMessage) error

// renderer returns the renderer of the global --json flag and the --ndjson flag of a
// listing, refusing both together.
func (a *App) renderer(ndjson bool) (renderer, error) {
	mode := output.Text
	switch {
	case a.global.json && ndjson:
		return renderer{}, failure.New(failure.Usage, "", "--json and --ndjson are exclusive, expected at most one")
	case a.global.json:
		mode = output.JSON
	case ndjson:
		mode = output.NDJSON
	}
	return renderer{mode: mode, printer: a.printer()}, nil
}

// list prints a JSON array body: unchanged with --json, one element per line with
// --ndjson, and as a table with a header otherwise. It streams and closes body.
func (r renderer) list(body io.ReadCloser, columns []output.Column, row rowFunc) error {
	switch r.mode {
	case output.JSON:
		return r.passthrough(body)
	case output.NDJSON:
		return r.printer.NDJSON(client.Elements(body))
	}
	return r.table(body, columns, row)
}

func (r renderer) table(body io.ReadCloser, columns []output.Column, row rowFunc) error {
	table := r.printer.Table(columns...)
	if err := table.Header(); err != nil {
		_ = body.Close()
		return err
	}
	for raw, err := range client.Elements(body) {
		if err != nil {
			return err
		}
		if err := tableRow(table, raw, row); err != nil {
			return err
		}
	}
	return nil
}

func tableRow(table *output.Table, raw json.RawMessage, row rowFunc) error {
	cells, err := row(raw)
	if err != nil {
		return err
	}
	return table.Row(cells...)
}

// object prints a JSON object body: unchanged with --json, compacted on one line with
// --ndjson, and through render otherwise. It closes body.
func (r renderer) object(body io.ReadCloser, render objectFunc) error {
	if r.mode == output.JSON {
		return r.passthrough(body)
	}
	raw, err := readObject(body)
	if err != nil {
		return err
	}
	if r.mode == output.NDJSON {
		return r.printer.NDJSON(single(raw))
	}
	return render(r.printer, raw)
}

func (r renderer) passthrough(body io.ReadCloser) error {
	defer func() { _ = body.Close() }()
	return r.printer.JSON(body)
}

func readObject(body io.ReadCloser) (json.RawMessage, error) {
	defer func() { _ = body.Close() }()
	raw, err := io.ReadAll(body)
	if err != nil {
		return nil, failure.Wrap(failure.Unavailable, "reading the server response", err)
	}
	return raw, nil
}

func single(raw json.RawMessage) iter.Seq2[json.RawMessage, error] {
	return func(yield func(json.RawMessage, error) bool) { yield(raw, nil) }
}

// malformedRecord reports a record that does not decode into its schema type.
func malformedRecord(raw json.RawMessage, err error) error {
	return failure.Wrap(failure.Internal, fmt.Sprintf("malformed server record %.200s", raw), err)
}
