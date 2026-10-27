package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/markupp-labs/markupp-cli/internal/failure"
	"github.com/markupp-labs/markupp-cli/internal/output"
	"github.com/markupp-labs/markupp-cli/internal/output/outputtest"
	"github.com/markupp-labs/markupp-cli/internal/schema"
)

const notesBody = `[{"id":"n1","title":"alpha","path":"a.md"},{"id":"n2","title":"be\u001bta","path":"b/b.md"}]`

var summaryColumns = []output.Column{{Name: "ID", Width: 4}, {Name: "PATH", Width: 8}, {Name: "TITLE"}}

func summaryRow(raw json.RawMessage) ([]string, error) {
	var note schema.NoteSummary
	if err := json.Unmarshal(raw, &note); err != nil {
		return nil, malformedRecord(raw, err)
	}
	return []string{note.ID, note.Path, note.Title}, nil
}

func renderNamed(p *output.Printer, raw json.RawMessage) error {
	var note schema.NoteSummary
	if err := json.Unmarshal(raw, &note); err != nil {
		return malformedRecord(raw, err)
	}
	return p.Fields(output.Field{Key: "id", Value: note.ID}, output.Field{Key: "path", Value: note.Path})
}

func body(text string) io.ReadCloser { return io.NopCloser(strings.NewReader(text)) }

func renderWith(t *testing.T, mode output.Mode, render func(renderer) error) string {
	t.Helper()
	var out bytes.Buffer
	if err := render(renderer{mode: mode, printer: output.NewPrinter(&out)}); err != nil {
		t.Fatalf("render in mode %d: %v", mode, err)
	}
	return out.String()
}

func TestListRendersEveryMode(t *testing.T) {
	for mode, name := range map[output.Mode]string{output.Text: "list_text", output.JSON: "list_json", output.NDJSON: "list_ndjson"} {
		got := renderWith(t, mode, func(r renderer) error {
			return r.list(body(notesBody), summaryColumns, summaryRow)
		})
		outputtest.Golden(t, name, []byte(got))
	}
}

func TestObjectRendersEveryMode(t *testing.T) {
	const note = "{\n  \"id\": \"n1\",\n  \"title\": \"alpha\",\n  \"path\": \"a.md\"\n}"
	for mode, name := range map[output.Mode]string{output.Text: "object_text", output.JSON: "object_json", output.NDJSON: "object_ndjson"} {
		got := renderWith(t, mode, func(r renderer) error { return r.object(body(note), renderNamed) })
		outputtest.Golden(t, name, []byte(got))
	}
}

func TestListStopsOnAMalformedRecord(t *testing.T) {
	var out bytes.Buffer
	r := renderer{mode: output.Text, printer: output.NewPrinter(&out)}
	err := r.list(body(`[{"id":1}]`), summaryColumns, summaryRow)
	if failure.KindOf(err) != failure.Internal || !strings.Contains(fmt.Sprint(err), `{"id":1}`) {
		t.Errorf("list error = %v, want an internal error naming the record", err)
	}
}

func TestRendererModeFromFlags(t *testing.T) {
	cases := []struct {
		json, ndjson bool
		want         output.Mode
	}{
		{false, false, output.Text}, {true, false, output.JSON}, {false, true, output.NDJSON},
	}
	for _, tc := range cases {
		app := newApp(newTestMu().deps(&bytes.Buffer{}, &bytes.Buffer{}))
		app.global.json = tc.json
		r, err := app.renderer(tc.ndjson)
		if err != nil || r.mode != tc.want {
			t.Errorf("json=%v ndjson=%v: mode %d, %v, want %d", tc.json, tc.ndjson, r.mode, err, tc.want)
		}
	}
}

func TestRendererRefusesJSONWithNDJSON(t *testing.T) {
	app := newApp(newTestMu().deps(&bytes.Buffer{}, &bytes.Buffer{}))
	app.global.json = true
	if _, err := app.renderer(true); failure.KindOf(err) != failure.Usage {
		t.Errorf("--json with --ndjson error = %v, want usage", err)
	}
}

// interruptedBody serves prefix, then fails as a body read does after its context
// was canceled.
type interruptedBody struct {
	prefix *strings.Reader
}

func (b *interruptedBody) Read(p []byte) (int, error) {
	if b.prefix.Len() > 0 {
		return b.prefix.Read(p)
	}
	return 0, fmt.Errorf("read tcp 127.0.0.1:80: %w", context.Canceled)
}

func (b *interruptedBody) Close() error { return nil }

func TestObjectInterruptedWhileReadingIsAnInterrupt(t *testing.T) {
	var out bytes.Buffer
	r := renderer{mode: output.Text, printer: output.NewPrinter(&out)}
	err := r.object(&interruptedBody{prefix: strings.NewReader(`{"id":`)}, renderNamed)
	if got := failure.KindOf(err); got != failure.Interrupted {
		t.Errorf("interrupted read error = %v (kind %s), want interrupted", err, got)
	}
}
